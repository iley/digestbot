package digest

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/iley/digestbot/internal/segment"
)

// NamedSegment pairs a segment with a human-readable name used to label it in
// the digest when it fails.
type NamedSegment struct {
	Name    string
	Segment segment.Segment
}

// Compose runs each segment sequentially and joins the results with blank
// lines. A segment that fails is replaced by a brief error notice (and the
// underlying error is logged) so the digest is always produced.
func Compose(ctx context.Context, segments []NamedSegment) string {
	parts := make([]string, 0, len(segments))
	for _, s := range segments {
		text, err := s.Segment.Produce(ctx)
		if err != nil {
			log.Printf("segment %q failed: %v", s.Name, err)
			text = fmt.Sprintf("<b>%s</b>\n⚠️ Failed to load.", segment.EscapeHTML(s.Name))
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n\n")
}
