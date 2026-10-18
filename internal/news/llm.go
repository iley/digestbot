package news

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/iley/digestbot/internal/llm"
)

// LLMRanker selects stories using a text completion model.
type LLMRanker struct{ LLM llm.LLM }

func (r *LLMRanker) Rank(ctx context.Context, articles []Article, limit int) ([]int, error) {
	count := min(limit, len(articles))
	if count <= 0 {
		return nil, nil
	}
	started := time.Now()
	prompt := fmt.Sprintf("Here are news headlines and their RSS descriptions:\n%s\nPick the %d most important news stories, in descending order of importance. Return ONLY a JSON array of %d unique 1-based article numbers, for example [3,1,2]. Do not summarize the articles.", articleList(articles), count, count)
	response, err := r.LLM.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("rank articles: %w", err)
	}
	var indices []int
	if err := json.Unmarshal([]byte(response), &indices); err != nil {
		return nil, fmt.Errorf("parse ranking: %w", err)
	}
	for i := range indices {
		indices[i]--
	}
	if err := ValidateRanking(indices, len(articles), count); err != nil {
		return nil, err
	}
	log.Printf("news ranking succeeded provider=llm candidates=%d selected=%d duration=%s", len(articles), count, time.Since(started))
	return indices, nil
}

// LLMSummarizer summarizes only the supplied stories, without selecting or reordering them.
type LLMSummarizer struct{ LLM llm.LLM }

func (s *LLMSummarizer) Summarize(ctx context.Context, articles []Article, language string) ([]string, error) {
	if len(articles) == 0 {
		return nil, nil
	}
	started := time.Now()
	if language == "" {
		language = "en"
	}
	prompt := fmt.Sprintf("Summarize each of these news stories in one sentence, using only its headline and RSS description. Output language: %s. Preserve article order; do not select, omit, or reorder stories. Return ONLY a JSON array of %d non-empty strings, one summary per article.\n%s", language, len(articles), articleList(articles))
	response, err := s.LLM.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("summarize articles: %w", err)
	}
	var summaries []string
	if err := json.Unmarshal([]byte(response), &summaries); err != nil {
		return nil, fmt.Errorf("parse summaries: %w", err)
	}
	if err := ValidateSummaries(summaries, len(articles)); err != nil {
		return nil, err
	}
	log.Printf("news summarization succeeded provider=llm articles=%d duration=%s", len(articles), time.Since(started))
	return summaries, nil
}

func articleList(articles []Article) string {
	var b strings.Builder
	for i, a := range articles {
		fmt.Fprintf(&b, "%d. %s", i+1, a.Title)
		if a.Summary != "" {
			fmt.Fprintf(&b, " — %s", a.Summary)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
