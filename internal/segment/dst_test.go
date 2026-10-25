package segment

import (
	"context"
	"testing"
	"time"
)

func TestDST(t *testing.T) {
	for _, tc := range []struct {
		name, timezone, now, want string
	}{
		{"default spring", "", "2026-03-29T08:00:00Z", "<b>⏰ Clocks moved one hour forward</b>"},
		{"autumn", "Europe/Dublin", "2026-10-25T08:00:00Z", "<b>⏰ Clocks moved one hour back</b>"},
		{"before transition", "Europe/Dublin", "2026-03-29T00:59:59Z", ""},
		{"at transition", "Europe/Dublin", "2026-03-29T01:00:00Z", "<b>⏰ Clocks moved one hour forward</b>"},
		{"just inside window", "Europe/Dublin", "2026-03-30T00:59:59Z", "<b>⏰ Clocks moved one hour forward</b>"},
		{"expired window", "Europe/Dublin", "2026-03-30T01:00:00Z", ""},
		{"ordinary day", "Europe/Dublin", "2026-07-01T08:00:00Z", ""},
		{"no DST", "UTC", "2026-03-29T08:00:00Z", ""},
		{"different transition date", "America/New_York", "2026-03-08T12:00:00Z", "<b>⏰ Clocks moved one hour forward</b>"},
		{"Dublin unchanged on US date", "Europe/Dublin", "2026-03-08T12:00:00Z", ""},
		{"half hour back", "Australia/Lord_Howe", "2026-04-05T00:00:00Z", "<b>⏰ Clocks moved 30 minutes back</b>"},
		{"half hour forward", "Australia/Lord_Howe", "2026-10-04T00:00:00Z", "<b>⏰ Clocks moved 30 minutes forward</b>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, tc.now)
			if err != nil {
				t.Fatal(err)
			}
			d := &DST{Timezone: tc.timezone, now: func() time.Time { return now }}
			got, err := d.Produce(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDSTInvalidTimezone(t *testing.T) {
	d := &DST{Timezone: "Invalid/Timezone"}
	if _, err := d.Produce(context.Background()); err == nil {
		t.Fatal("expected invalid timezone error")
	}
}
