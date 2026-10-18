package segment

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/iley/digestbot/internal/news"
)

const maxNewsArticles = 5

// maxRankingArticles bounds the candidate pool sent to the ranker.
const maxRankingArticles = 15
const selectedNewsArticles = 3

type News struct {
	Title      string
	Fetcher    news.FeedFetcher
	Ranker     news.Ranker
	Summarizer news.Summarizer
	Language   string // "ru" for Russian summaries; default is English
}

func (n *News) Produce(ctx context.Context) (string, error) {
	articles, err := n.Fetcher.Fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("news %s: %w", n.Title, err)
	}

	if len(articles) == 0 {
		return fmt.Sprintf("<b>%s</b>\nNo articles available.", EscapeHTML(n.Title)), nil
	}

	if n.Ranker != nil || n.Summarizer != nil {
		if n.Ranker == nil || n.Summarizer == nil {
			return "", fmt.Errorf("news %s: ranker and summarizer must both be configured", n.Title)
		}
		return n.produceDigest(ctx, articles)
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

func (n *News) produceDigest(ctx context.Context, articles []news.Article) (string, error) {
	if len(articles) > maxRankingArticles {
		articles = articles[:maxRankingArticles]
	}
	count := min(selectedNewsArticles, len(articles))
	indices, err := n.Ranker.Rank(ctx, articles, count)
	if err == nil {
		err = news.ValidateRanking(indices, len(articles), count)
	}
	if err != nil {
		return "", fmt.Errorf("news %s ranking: %w", n.Title, err)
	}
	selected := make([]news.Article, len(indices))
	for i, index := range indices {
		selected[i] = articles[index]
	}
	summaries, err := n.Summarizer.Summarize(ctx, selected, n.Language)
	if err == nil {
		err = news.ValidateSummaries(summaries, len(selected))
	}
	if err != nil {
		return "", fmt.Errorf("news %s summarization: %w", n.Title, err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>", EscapeHTML(n.Title))
	for i, a := range selected {
		fmt.Fprintf(&b, "\n%d. %s <a href=\"%s\">Read more</a>", i+1, EscapeHTML(summaries[i]), sanitizeURL(a.Link))
	}
	return b.String(), nil
}

// sanitizeURL ensures a URL is safe for use in an HTML href attribute.
func sanitizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return EscapeHTML(u.String())
}
