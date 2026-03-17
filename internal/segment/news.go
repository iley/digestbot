package segment

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/iley/digestbot/internal/news"
)

const maxNewsArticles = 5

type News struct {
	Title   string
	Fetcher news.FeedFetcher
}

func (n *News) Produce(ctx context.Context) (string, error) {
	articles, err := n.Fetcher.Fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("news %s: %w", n.Title, err)
	}

	if len(articles) == 0 {
		return fmt.Sprintf("<b>%s</b>\nNo articles available.", EscapeHTML(n.Title)), nil
	}

	if len(articles) > maxNewsArticles {
		articles = articles[:maxNewsArticles]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<b>%s</b>", EscapeHTML(n.Title))
	for _, a := range articles {
		fmt.Fprintf(&b, "\n• <a href=\"%s\">%s</a>", sanitizeURL(a.Link), EscapeHTML(a.Title))
	}

	return b.String(), nil
}

// sanitizeURL ensures a URL is safe for use in an HTML href attribute.
// It re-parses and re-encodes so that characters like & and " can't break the HTML.
func sanitizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return EscapeHTML(u.String())
}
