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
