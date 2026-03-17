package segment

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/iley/digestbot/internal/news"
)

// mockFetcher returns a fixed list of articles.
type mockFetcher struct {
	articles []news.Article
}

func (m *mockFetcher) Fetch(_ context.Context) ([]news.Article, error) {
	return m.articles, nil
}

// mockLLM returns a fixed response.
type mockLLM struct {
	response string
	received string
}

func (m *mockLLM) Complete(_ context.Context, prompt string) (string, error) {
	m.received = prompt
	return m.response, nil
}

type failingLLM struct{}

func (f *failingLLM) Complete(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("llm error")
}

func TestNewsWithLLM(t *testing.T) {
	articles := []news.Article{
		{Title: "Article One", Summary: "Summary one", Link: "https://example.com/1"},
		{Title: "Article Two", Summary: "Summary two", Link: "https://example.com/2"},
		{Title: "Article Three", Summary: "Summary three", Link: "https://example.com/3"},
	}

	llm := &mockLLM{
		response: `[{"index": 1, "summary": "First is important"}, {"index": 3, "summary": "Third matters"}]`,
	}

	n := &News{
		Title:   "Test News",
		Fetcher: &mockFetcher{articles: articles},
		LLM:     llm,
	}

	result, err := n.Produce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain the title
	if !strings.Contains(result, "<b>Test News</b>") {
		t.Errorf("missing title in result: %s", result)
	}

	// Should contain summaries with links
	if !strings.Contains(result, "First is important") {
		t.Errorf("missing first summary in result: %s", result)
	}
	if !strings.Contains(result, "Third matters") {
		t.Errorf("missing third summary in result: %s", result)
	}
	if !strings.Contains(result, "https://example.com/1") {
		t.Errorf("missing first link in result: %s", result)
	}
	if !strings.Contains(result, "https://example.com/3") {
		t.Errorf("missing third link in result: %s", result)
	}

	// Prompt should mention article titles
	if !strings.Contains(llm.received, "Article One") {
		t.Errorf("prompt should contain article titles: %s", llm.received)
	}
}

func TestNewsWithLLMRussian(t *testing.T) {
	articles := []news.Article{
		{Title: "Статья один", Summary: "Описание один", Link: "https://example.com/1"},
		{Title: "Статья два", Summary: "Описание два", Link: "https://example.com/2"},
		{Title: "Статья три", Summary: "Описание три", Link: "https://example.com/3"},
	}

	llm := &mockLLM{
		response: `[{"index": 1, "summary": "Первая важная"}, {"index": 2, "summary": "Вторая важная"}]`,
	}

	n := &News{
		Title:    "Meduza",
		Fetcher:  &mockFetcher{articles: articles},
		LLM:      llm,
		Language: "ru",
	}

	result, err := n.Produce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<b>Meduza</b>") {
		t.Errorf("missing title in result: %s", result)
	}
	if !strings.Contains(result, "Первая важная") {
		t.Errorf("missing first summary in result: %s", result)
	}

	// Prompt should be in Russian
	if !strings.Contains(llm.received, "Выбери 3 самые важные") {
		t.Errorf("prompt should be in Russian: %s", llm.received)
	}
}

func TestNewsWithoutLLM(t *testing.T) {
	articles := []news.Article{
		{Title: "Headline A", Link: "https://example.com/a"},
		{Title: "Headline B", Link: "https://example.com/b"},
	}

	n := &News{
		Title:   "Raw News",
		Fetcher: &mockFetcher{articles: articles},
	}

	result, err := n.Produce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Headline A") {
		t.Errorf("missing headline in result: %s", result)
	}
	if !strings.Contains(result, "Headline B") {
		t.Errorf("missing headline in result: %s", result)
	}
}

func TestNewsLLMError(t *testing.T) {
	articles := []news.Article{
		{Title: "Article", Summary: "Summary", Link: "https://example.com/1"},
	}

	n := &News{
		Title:   "Test",
		Fetcher: &mockFetcher{articles: articles},
		LLM:     &failingLLM{},
	}

	_, err := n.Produce(context.Background())
	if err == nil {
		t.Fatal("expected error from LLM failure, got nil")
	}
}
