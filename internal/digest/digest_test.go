package digest

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type stubSegment struct {
	text string
	err  error
}

func (s stubSegment) Produce(ctx context.Context) (string, error) {
	return s.text, s.err
}

func TestComposeJoinsSegments(t *testing.T) {
	out := Compose(context.Background(), []NamedSegment{
		{Name: "A", Segment: stubSegment{text: "first"}},
		{Name: "B", Segment: stubSegment{text: "second"}},
	})

	if out != "first\n\nsecond" {
		t.Errorf("got %q, want %q", out, "first\n\nsecond")
	}
}

func TestComposeOmitsEmptySegments(t *testing.T) {
	out := Compose(context.Background(), []NamedSegment{
		{Name: "DST", Segment: stubSegment{}},
		{Name: "A", Segment: stubSegment{text: "first"}},
		{Name: "Empty", Segment: stubSegment{text: " \n"}},
		{Name: "B", Segment: stubSegment{text: "second"}},
		{Name: "Empty", Segment: stubSegment{}},
	})
	if out != "first\n\nsecond" {
		t.Fatalf("got %q", out)
	}
	if out := Compose(context.Background(), []NamedSegment{{Name: "DST", Segment: stubSegment{}}}); out != "" {
		t.Fatalf("all-empty digest got %q", out)
	}
}

// A failing segment must not abort the digest; it is replaced by a notice
// while the other segments still render.
func TestComposeReplacesFailedSegment(t *testing.T) {
	out := Compose(context.Background(), []NamedSegment{
		{Name: "Weather", Segment: stubSegment{err: fmt.Errorf("boom")}},
		{Name: "News", Segment: stubSegment{text: "headline"}},
	})

	if !strings.Contains(out, "<b>Weather</b>") {
		t.Errorf("missing failed-segment header in %q", out)
	}
	if strings.Contains(out, "boom") {
		t.Errorf("raw error leaked into digest: %q", out)
	}
	if !strings.Contains(out, "headline") {
		t.Errorf("healthy segment dropped from %q", out)
	}
}
