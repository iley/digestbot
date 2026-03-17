package news

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type RSSFetcher struct {
	FeedURL string
	Client  *http.Client // optional, for testing
}

func (f *RSSFetcher) Fetch(ctx context.Context) ([]Article, error) {
	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.FeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing feed: %w", err)
	}

	articles := make([]Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		articles = append(articles, Article{
			Title:   item.Title,
			Summary: item.Description,
			Link:    item.Link,
		})
	}

	return articles, nil
}
