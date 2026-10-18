package news

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEditorialSuccessLogs(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	articles := []Article{{Title: "private article", Summary: "private description"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"model":"jev-resolved","answers":{"article_0":{"type":"score","score":2}}}`)
	}))
	defer server.Close()
	ranker := &JevRanker{APIKey: "private-key", BaseURL: server.URL, Client: server.Client()}
	if _, err := ranker.Rank(context.Background(), articles, 1); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"news ranking succeeded provider=jev", `requested_model="jev-latest"`, `model="jev-resolved"`, "candidates=1 selected=1 duration="} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q: %s", want, output.String())
		}
	}
	client := completionFunc(func(context.Context, string) (string, error) { return `[1]`, nil })
	if _, err := (&LLMRanker{LLM: client}).Rank(context.Background(), articles, 1); err != nil {
		t.Fatal(err)
	}
	client = completionFunc(func(context.Context, string) (string, error) { return `["private summary"]`, nil })
	if _, err := (&LLMSummarizer{LLM: client}).Summarize(context.Background(), articles, ""); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"news ranking succeeded provider=llm candidates=1 selected=1 duration=", "news summarization succeeded provider=llm articles=1 duration="} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q: %s", want, output.String())
		}
	}
	if strings.Contains(output.String(), "private") {
		t.Fatalf("logs contain private data: %s", output.String())
	}
}

func TestEditorialFailuresDoNotLogSuccess(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	articles := []Article{{Title: "A"}}
	client := completionFunc(func(context.Context, string) (string, error) { return `[]`, nil })
	if _, err := (&LLMRanker{LLM: client}).Rank(context.Background(), articles, 1); err == nil {
		t.Fatal("expected invalid ranking")
	}
	if _, err := (&LLMSummarizer{LLM: client}).Summarize(context.Background(), articles, ""); err == nil {
		t.Fatal("expected invalid summary")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"answers":{}}`) }))
	defer server.Close()
	ranker := &JevRanker{APIKey: "key", BaseURL: server.URL, Client: server.Client()}
	if _, err := ranker.Rank(context.Background(), articles, 1); err == nil {
		t.Fatal("expected invalid Jev response")
	}
	if _, err := ranker.Rank(context.Background(), nil, 1); err != nil {
		t.Fatal(err)
	}
	if output.Len() != 0 {
		t.Fatalf("unexpected success logs: %s", output.String())
	}
}
