package news

import (
	"context"
	"fmt"
	"strings"
)

// Ranker selects up to limit articles in descending order of importance.
// Returned indices are zero-based, unique, and refer to the input slice.
// It must return min(limit, len(articles)) indices for a positive limit.
// A non-positive limit returns no selections. Implementations must not mutate articles.
type Ranker interface {
	Rank(ctx context.Context, articles []Article, limit int) ([]int, error)
}

// Summarizer writes one summary per article in the requested language.
// Results must have the same length and order as articles, with no blank summaries.
// Empty language means English. Implementations must not mutate articles.
type Summarizer interface {
	Summarize(ctx context.Context, articles []Article, language string) ([]string, error)
}

// ValidateRanking checks the Ranker contract before indices are used.
func ValidateRanking(indices []int, articleCount, count int) error {
	if len(indices) != count {
		return fmt.Errorf("ranking returned %d articles, want %d", len(indices), count)
	}
	seen := make(map[int]bool, len(indices))
	for _, index := range indices {
		if index < 0 || index >= articleCount {
			return fmt.Errorf("ranking index %d out of range", index)
		}
		if seen[index] {
			return fmt.Errorf("ranking contains duplicate index %d", index)
		}
		seen[index] = true
	}
	return nil
}

// ValidateSummaries checks the Summarizer contract before rendering.
func ValidateSummaries(summaries []string, count int) error {
	if len(summaries) != count {
		return fmt.Errorf("summarizer returned %d summaries, want %d", len(summaries), count)
	}
	for i, summary := range summaries {
		if strings.TrimSpace(summary) == "" {
			return fmt.Errorf("summary %d is empty", i)
		}
	}
	return nil
}
