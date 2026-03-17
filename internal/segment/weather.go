package segment

import (
	"context"
	"fmt"

	"github.com/iley/digestbot/internal/weather"
)

type Weather struct {
	Provider weather.WeatherProvider
}

func (w *Weather) Produce(ctx context.Context) (string, error) {
	f, err := w.Provider.Today(ctx)
	if err != nil {
		return "", fmt.Errorf("weather: %w", err)
	}

	desc := weather.DescribeWeatherCode(f.WeatherCode)
	return fmt.Sprintf(
		"<b>Weather</b>\n%s, %.0f–%.0f °C, precipitation %.1f mm",
		EscapeHTML(desc), f.TemperatureMin, f.TemperatureMax, f.Precipitation,
	), nil
}
