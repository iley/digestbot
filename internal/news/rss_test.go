package news

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Irish Times</title>
    <item>
      <title>Budget announced</title>
      <description>The government has unveiled its new budget plan.</description>
      <link>https://example.com/budget</link>
    </item>
    <item>
      <title>Weather warning issued</title>
      <description>Met Éireann issues orange weather warning.</description>
      <link>https://example.com/weather</link>
    </item>
  </channel>
</rss>`

const emptyRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Irish Times</title>
  </channel>
</rss>`

func TestRSSFetcherSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(sampleRSS))
	}))
	defer srv.Close()

	fetcher := &RSSFetcher{FeedURL: srv.URL}
	articles, err := fetcher.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("got %d articles, want 2", len(articles))
	}

	if articles[0].Title != "Budget announced" {
		t.Errorf("Title = %q, want %q", articles[0].Title, "Budget announced")
	}
	if articles[0].Summary != "The government has unveiled its new budget plan." {
		t.Errorf("Summary = %q, want %q", articles[0].Summary, "The government has unveiled its new budget plan.")
	}
	if articles[0].Link != "https://example.com/budget" {
		t.Errorf("Link = %q, want %q", articles[0].Link, "https://example.com/budget")
	}

	if articles[1].Title != "Weather warning issued" {
		t.Errorf("Title = %q, want %q", articles[1].Title, "Weather warning issued")
	}
}

func TestRSSFetcherServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	fetcher := &RSSFetcher{FeedURL: srv.URL}
	_, err := fetcher.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestRSSFetcherEmptyFeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(emptyRSS))
	}))
	defer srv.Close()

	fetcher := &RSSFetcher{FeedURL: srv.URL}
	articles, err := fetcher.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 0 {
		t.Errorf("got %d articles, want 0", len(articles))
	}
}
