package segment

import (
	"context"
	"html"
)

// Segment produces a piece of the daily digest.
// Produce must return valid Telegram HTML.
type Segment interface {
	Produce(ctx context.Context) (string, error)
}

// EscapeHTML escapes text for safe inclusion in Telegram HTML messages.
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}
