package news

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestJevRankerRequestAndRanking(t *testing.T) {
	for _, model := range []string{"", "jev-pinned"} {
		t.Run(model, func(t *testing.T) {
			calls := 0
			articles := []Article{{Title: "Irish story", Summary: "Details", Link: "https://example.com"}, {Title: "Русские новости", Summary: "Описание"}, {Title: "Third"}, {Title: "Fourth"}}
			original := append([]Article(nil), articles...)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "POST" || r.URL.Path != "/v1/systemone" || r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var request jevRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				expectedModel := model
				if expectedModel == "" {
					expectedModel = "jev-latest"
				}
				if request.Model != expectedModel || len(request.State) != 4 || len(request.Questions) != 4 {
					t.Errorf("unexpected payload: %+v", request)
				}
				for i, a := range request.State {
					id := fmt.Sprintf("article_%d", i)
					q := request.Questions[id]
					if a.ID != id || a.Title != articles[i].Title || a.Summary != articles[i].Summary || q.Type != "score" || !strings.Contains(q.Instructions, id) || len(q.Criteria) != 5 {
						t.Errorf("incorrect article/question: %+v %+v", a, q)
					}
				}
				fmt.Fprint(w, `{"model":"jev-test","answers":{"article_3":{"type":"score","score":0},"article_2":{"type":"score","score":3.5},"article_0":{"type":"score","score":1.2},"article_1":{"type":"score","score":3.5}},"usage":{"input_tokens":100}}`)
			}))
			defer server.Close()
			ranker := &JevRanker{APIKey: "test-key", Model: model, BaseURL: server.URL + "/v1/", Client: server.Client()}
			for _, tc := range []struct {
				limit int
				want  []int
			}{{3, []int{1, 2, 0}}, {10, []int{1, 2, 0, 3}}, {1, []int{1}}} {
				got, err := ranker.Rank(context.Background(), articles, tc.limit)
				if err != nil || !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("%v, %v; want %v", got, err, tc.want)
				}
			}
			if calls != 3 || !reflect.DeepEqual(articles, original) {
				t.Fatal("unexpected request count or mutated input")
			}
		})
	}
}

func TestJevRankerInvalidResponses(t *testing.T) {
	for _, body := range []string{
		`not json`, `null`, `{}`, `{"answers":{}}`,
		`{"answers":{"article_0":{"type":"score"}}}`,
		`{"answers":{"article_0":{"type":"score","score":null}}}`,
		`{"answers":{"article_0":{"type":"choice","score":1}}}`,
		`{"answers":{"article_0":{"type":"score","score":-1}}}`,
		`{"answers":{"article_0":{"type":"score","score":4.1}}}`,
		`{"answers":{"article_0":{"type":"score","score":"2"}}}`,
		`{"answers":{"article_0":{"type":"score","score":1e400}}}`,
		`{"answers":{"wrong_id":{"type":"score","score":2}}}`,
		`{"answers":{"article_0":{"type":"score","score":2}}} trailing`,
		strings.Repeat(" ", (1<<20)+1),
	} {
		t.Run(fmt.Sprintf("case_%d", len(body)), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
			defer server.Close()
			ranker := &JevRanker{APIKey: "key", BaseURL: server.URL, Client: server.Client()}
			if _, err := ranker.Rank(context.Background(), []Article{{Title: "A"}}, 1); err == nil {
				t.Fatal("expected response validation error")
			}
		})
	}
}

func TestJevRankerHTTPError(t *testing.T) {
	for _, status := range []int{401, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				fmt.Fprint(w, "private upstream detail")
			}))
			defer server.Close()
			ranker := &JevRanker{APIKey: "key", BaseURL: server.URL, Client: server.Client()}
			_, err := ranker.Rank(context.Background(), []Article{{Title: "A"}}, 1)
			if err == nil || !strings.Contains(err.Error(), fmt.Sprint(status)) || strings.Contains(err.Error(), "private upstream detail") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

type jevTransportFunc func(*http.Request) (*http.Response, error)

func (f jevTransportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestJevRankerEmptyKeyAndTransport(t *testing.T) {
	ranker := &JevRanker{}
	for _, limit := range []int{-1, 0, 1} {
		if got, err := ranker.Rank(context.Background(), nil, limit); err != nil || len(got) != 0 {
			t.Fatalf("%v, %v", got, err)
		}
	}
	articles := []Article{{Title: "A"}}
	if got, err := ranker.Rank(context.Background(), articles, 0); err != nil || len(got) != 0 {
		t.Fatalf("%v, %v", got, err)
	}
	if _, err := ranker.Rank(context.Background(), articles, 1); err == nil {
		t.Fatal("expected missing key error")
	}
	sentinel := errors.New("network failure")
	ranker.APIKey = "key"
	ranker.Client = &http.Client{Transport: jevTransportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.typesafe.ai/v1/systemone" {
			t.Errorf("wrong default URL: %s", r.URL)
		}
		if _, ok := r.Context().Deadline(); !ok {
			t.Error("missing deadline")
		}
		if r.Context().Err() != nil {
			return nil, r.Context().Err()
		}
		return nil, sentinel
	})}
	if _, err := ranker.Rank(context.Background(), articles, 1); !errors.Is(err, sentinel) {
		t.Fatalf("%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ranker.Rank(ctx, articles, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("%v", err)
	}
	ranker.BaseURL = ":bad"
	if _, err := ranker.Rank(context.Background(), articles, 1); err == nil {
		t.Fatal("expected URL error")
	}
}
