package segment

import (
	"context"
	"fmt"
)

// Placeholder is a test segment that returns static content.
type Placeholder struct {
	Title string
	Body  string
}

func (p *Placeholder) Produce(ctx context.Context) (string, error) {
	return fmt.Sprintf("<b>%s</b>\n%s", EscapeHTML(p.Title), EscapeHTML(p.Body)), nil
}
