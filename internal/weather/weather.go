package weather

import "context"

type Forecast struct {
	TemperatureMin float64
	TemperatureMax float64
	Precipitation  float64 // mm
	WeatherCode    int     // WMO code
}

type WeatherProvider interface {
	Today(ctx context.Context) (*Forecast, error)
}
