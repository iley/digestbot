package news

import "context"

type Article struct {
	Title   string
	Summary string
	Link    string
}

// FeedFetcher retrieves articles from a news source.
type FeedFetcher interface {
	Fetch(ctx context.Context) ([]Article, error)
}
