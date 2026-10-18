package segment

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/iley/digestbot/internal/news"
)

type mockFetcher struct {
	articles []news.Article
	err      error
}

func (m *mockFetcher) Fetch(context.Context) ([]news.Article, error) { return m.articles, m.err }

type rankFunc func(context.Context, []news.Article, int) ([]int, error)

func (f rankFunc) Rank(ctx context.Context, a []news.Article, n int) ([]int, error) {
	return f(ctx, a, n)
}

type summarizeFunc func(context.Context, []news.Article, string) ([]string, error)

func (f summarizeFunc) Summarize(ctx context.Context, a []news.Article, l string) ([]string, error) {
	return f(ctx, a, l)
}

func TestNewsPipeline(t *testing.T) {
	for _, language := range []string{"", "ru"} {
		t.Run(language, func(t *testing.T) {
			articles := make([]news.Article, 20)
			for i := range articles {
				articles[i] = news.Article{Title: fmt.Sprintf("Story %d", i), Link: fmt.Sprintf("https://example.com/%d", i)}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ranked := false
			n := &News{Title: "News & More", Fetcher: &mockFetcher{articles: articles}, Language: language,
				Ranker: rankFunc(func(got context.Context, a []news.Article, count int) ([]int, error) {
					if got != ctx || count != 3 || !reflect.DeepEqual(a, articles[:15]) {
						t.Fatalf("unexpected ranking input: %v, %d", a, count)
					}
					ranked = true
					return []int{4, 0, 2}, nil
				}),
				Summarizer: summarizeFunc(func(got context.Context, a []news.Article, lang string) ([]string, error) {
					if !ranked || got != ctx || lang != language || !reflect.DeepEqual(a, []news.Article{articles[4], articles[0], articles[2]}) {
						t.Fatalf("unexpected summarization input: %v, %q", a, lang)
					}
					return []string{"First <story>", "Second", "Third"}, nil
				})}
			result, err := n.Produce(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := "<b>News &amp; More</b>\n1. First &lt;story&gt; <a href=\"https://example.com/4\">Read more</a>\n2. Second <a href=\"https://example.com/0\">Read more</a>\n3. Third <a href=\"https://example.com/2\">Read more</a>"
			if result != want {
				t.Fatalf("got %s, want %s", result, want)
			}
		})
	}
}

func TestNewsFailures(t *testing.T) {
	sentinel := errors.New("provider failed")
	for _, tc := range []struct {
		name       string
		indices    []int
		rankErr    error
		summaries  []string
		summaryErr error
		stage      string
	}{
		{name: "ranking error", rankErr: sentinel, stage: "ranking"},
		{name: "duplicate", indices: []int{0, 0}, stage: "ranking"},
		{name: "negative", indices: []int{-1, 0}, stage: "ranking"},
		{name: "out of range", indices: []int{0, 2}, stage: "ranking"},
		{name: "missing selection", indices: []int{0}, stage: "ranking"},
		{name: "summary error", indices: []int{1, 0}, summaryErr: sentinel, stage: "summarization"},
		{name: "missing summary", indices: []int{1, 0}, summaries: []string{"one"}, stage: "summarization"},
		{name: "blank summary", indices: []int{1, 0}, summaries: []string{"one", " "}, stage: "summarization"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			n := &News{Title: "Test", Fetcher: &mockFetcher{articles: []news.Article{{Title: "A"}, {Title: "B"}}},
				Ranker: rankFunc(func(_ context.Context, _ []news.Article, count int) ([]int, error) {
					if count != 2 {
						t.Fatalf("count = %d", count)
					}
					return tc.indices, tc.rankErr
				}),
				Summarizer: summarizeFunc(func(context.Context, []news.Article, string) ([]string, error) {
					called = true
					return tc.summaries, tc.summaryErr
				})}
			_, err := n.Produce(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.stage) {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.stage == "ranking" && called {
				t.Fatal("summarized after failed ranking")
			}
			if (tc.rankErr != nil || tc.summaryErr != nil) && !errors.Is(err, sentinel) {
				t.Fatal("lost original error")
			}
		})
	}
}

func TestNewsRawAndEmpty(t *testing.T) {
	n := &News{Title: "Raw", Fetcher: &mockFetcher{articles: []news.Article{{Title: "A & B", Link: "https://example.com/a?x=1&y=2"}, {Title: "B"}, {Title: "C"}, {Title: "D"}, {Title: "E"}, {Title: "Sixth"}}}}
	result, err := n.Produce(context.Background())
	if err != nil || !strings.Contains(result, "A &amp; B") || strings.Contains(result, "Sixth") || !strings.Contains(result, "x=1&amp;y=2") {
		t.Fatalf("%s, %v", result, err)
	}
	n.Fetcher = &mockFetcher{}
	result, err = n.Produce(context.Background())
	if err != nil || result != "<b>Raw</b>\nNo articles available." {
		t.Fatalf("%s, %v", result, err)
	}
}

func TestNewsMissingComponentAndFetchError(t *testing.T) {
	ranker := rankFunc(func(context.Context, []news.Article, int) ([]int, error) { t.Fatal("unexpected call"); return nil, nil })
	summarizer := summarizeFunc(func(context.Context, []news.Article, string) ([]string, error) {
		t.Fatal("unexpected call")
		return nil, nil
	})
	for _, n := range []*News{
		{Ranker: ranker}, {Summarizer: summarizer},
	} {
		n.Fetcher = &mockFetcher{articles: []news.Article{{Title: "A"}}}
		if _, err := n.Produce(context.Background()); err == nil {
			t.Fatal("expected configuration error")
		}
	}
	sentinel := errors.New("fetch failed")
	n := &News{Fetcher: &mockFetcher{err: sentinel}, Ranker: ranker, Summarizer: summarizer}
	if _, err := n.Produce(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("error: %v", err)
	}
}
