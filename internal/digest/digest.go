package digest

import (
	"context"
	"strings"

	"github.com/iley/digestbot/internal/segment"
)

// Compose runs each segment sequentially and joins results with blank lines.
func Compose(ctx context.Context, segments []segment.Segment) (string, error) {
	parts := make([]string, 0, len(segments))
	for _, seg := range segments {
		text, err := seg.Produce(ctx)
		if err != nil {
			return "", err
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n\n"), nil
}
