package segment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/iley/digestbot/internal/llm"
	"github.com/iley/digestbot/internal/news"
)

const maxNewsArticles = 5

type News struct {
	Title   string
	Fetcher news.FeedFetcher
	LLM     llm.LLM // nil = raw headlines (no summarization)
}

func (n *News) Produce(ctx context.Context) (string, error) {
	articles, err := n.Fetcher.Fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("news %s: %w", n.Title, err)
	}

	if len(articles) == 0 {
		return fmt.Sprintf("<b>%s</b>\nNo articles available.", EscapeHTML(n.Title)), nil
	}

	if n.LLM != nil {
		return n.produceWithLLM(ctx, articles)
	}
	return n.produceRaw(articles), nil
}

func (n *News) produceRaw(articles []news.Article) string {
	if len(articles) > maxNewsArticles {
		articles = articles[:maxNewsArticles]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>", EscapeHTML(n.Title))
	for _, a := range articles {
		fmt.Fprintf(&b, "\n• <a href=\"%s\">%s</a>", sanitizeURL(a.Link), EscapeHTML(a.Title))
	}
	return b.String()
}

type llmPick struct {
	Index   int    `json:"index"`
	Summary string `json:"summary"`
}

func (n *News) produceWithLLM(ctx context.Context, articles []news.Article) (string, error) {
	prompt := buildPrompt(n.Title, articles)

	response, err := n.LLM.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("news %s llm: %w", n.Title, err)
	}

	var picks []llmPick
	if err := json.Unmarshal([]byte(response), &picks); err != nil {
		return "", fmt.Errorf("news %s: failed to parse LLM response: %w", n.Title, err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>", EscapeHTML(n.Title))
	for i, p := range picks {
		if p.Index < 1 || p.Index > len(articles) {
			continue
		}
		a := articles[p.Index-1]
		fmt.Fprintf(&b, "\n%d. <a href=\"%s\">%s</a>\n%s",
			i+1, sanitizeURL(a.Link), EscapeHTML(a.Title), EscapeHTML(p.Summary))
	}
	return b.String(), nil
}

func buildPrompt(source string, articles []news.Article) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Here are today's headlines from %s:\n", source)
	for i, a := range articles {
		fmt.Fprintf(&b, "%d. %s", i+1, a.Title)
		if a.Summary != "" {
			fmt.Fprintf(&b, " — %s", a.Summary)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nPick the 3 most important news stories. Return a JSON array:\n")
	b.WriteString(`[{"index": <1-based>, "summary": "<1-2 sentence summary>"}]`)
	b.WriteString("\nReturn ONLY the JSON array, no other text.")
	return b.String()
}

// sanitizeURL ensures a URL is safe for use in an HTML href attribute.
func sanitizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return EscapeHTML(u.String())
}
