package segment

import (
	"context"
	"fmt"
	"time"
	_ "time/tzdata" // Keep timezone lookup working in minimal container images.
)

// DST announces clock changes in the previous 24 hours.
type DST struct {
	Timezone string
	now      func() time.Time
}

func (d *DST) Produce(ctx context.Context) (string, error) {
	timezone := d.Timezone
	if timezone == "" {
		timezone = "Europe/Dublin"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("dst: %w", err)
	}
	now := time.Now
	if d.now != nil {
		now = d.now
	}
	current := now().In(loc)
	_, before := current.Add(-24 * time.Hour).Zone()
	_, after := current.Zone()
	change := after - before
	if change == 0 {
		return "", nil
	}
	direction := "forward"
	if change < 0 {
		direction = "back"
		change = -change
	}
	amount := "one hour"
	if change != 3600 {
		amount = fmt.Sprintf("%d minutes", change/60)
	}
	return fmt.Sprintf("<b>⏰ Clocks moved %s %s</b>", amount, direction), nil
}
