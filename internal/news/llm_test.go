package news

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type completionFunc func(context.Context, string) (string, error)

func (f completionFunc) Complete(ctx context.Context, p string) (string, error) { return f(ctx, p) }

func TestLLMRanker(t *testing.T) {
	articles := []Article{{Title: "Alpha", Summary: "Details"}, {Title: "Beta"}, {Title: "Gamma"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &LLMRanker{LLM: completionFunc(func(got context.Context, p string) (string, error) {
		if got != ctx || !strings.Contains(p, "Alpha — Details") || !strings.Contains(p, "Pick the 2") || !strings.Contains(p, "Do not summarize") {
			t.Fatalf("unexpected prompt: %s", p)
		}
		return "[3,1]", nil
	})}
	indices, err := r.Rank(ctx, articles, 2)
	if err != nil || !reflect.DeepEqual(indices, []int{2, 0}) {
		t.Fatalf("%v, %v", indices, err)
	}
}

func TestLLMRankerInvalidResponses(t *testing.T) {
	for _, response := range []string{"not json", "null", "[]", "[1]", "[1,1]", "[0,2]", "[1,3]", "[1,2,3]", "[1.5,2]"} {
		t.Run(response, func(t *testing.T) {
			r := &LLMRanker{LLM: completionFunc(func(context.Context, string) (string, error) { return response, nil })}
			if _, err := r.Rank(context.Background(), []Article{{Title: "A"}, {Title: "B"}}, 3); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestLLMSummarizer(t *testing.T) {
	for _, language := range []string{"", "ru"} {
		t.Run(language, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			s := &LLMSummarizer{LLM: completionFunc(func(got context.Context, p string) (string, error) {
				expectedLanguage := language
				if expectedLanguage == "" {
					expectedLanguage = "en"
				}
				if got != ctx || !strings.Contains(p, "Output language: "+expectedLanguage) || !strings.Contains(p, "Alpha — Details") || !strings.Contains(p, "Preserve article order") {
					t.Fatalf("unexpected prompt: %s", p)
				}
				return `["First", "Вторая"]`, nil
			})}
			summaries, err := s.Summarize(ctx, []Article{{Title: "Alpha", Summary: "Details"}, {Title: "Beta"}}, language)
			if err != nil || !reflect.DeepEqual(summaries, []string{"First", "Вторая"}) {
				t.Fatalf("%v, %v", summaries, err)
			}
		})
	}
}

func TestLLMSummarizerInvalidResponses(t *testing.T) {
	for _, response := range []string{"not json", "null", "[]", `["one"]`, `["one","two","three"]`, `["one"," "]`, `["one",null]`, `[{"summary":"one"}]`} {
		t.Run(response, func(t *testing.T) {
			s := &LLMSummarizer{LLM: completionFunc(func(context.Context, string) (string, error) { return response, nil })}
			if _, err := s.Summarize(context.Background(), []Article{{Title: "A"}, {Title: "B"}}, ""); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestLLMEditorialEmptyAndFailures(t *testing.T) {
	sentinel := errors.New("completion failed")
	calls := 0
	client := completionFunc(func(context.Context, string) (string, error) { calls++; return "", sentinel })
	r := &LLMRanker{LLM: client}
	s := &LLMSummarizer{LLM: client}
	ctx := context.Background()
	for _, limit := range []int{-1, 0, 3} {
		if _, err := r.Rank(ctx, nil, limit); err != nil {
			t.Fatal(err)
		}
	}
	articles := []Article{{Title: "A"}}
	if _, err := r.Rank(ctx, articles, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Summarize(ctx, nil, ""); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("called LLM for empty work")
	}
	if _, err := r.Rank(ctx, articles, 3); !errors.Is(err, sentinel) {
		t.Fatalf("%v", err)
	}
	if _, err := s.Summarize(ctx, articles, ""); !errors.Is(err, sentinel) {
		t.Fatalf("%v", err)
	}
}
