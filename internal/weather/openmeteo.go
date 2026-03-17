package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type OpenMeteo struct {
	Latitude  float64
	Longitude float64
	Timezone  string
	BaseURL   string // override for testing; empty means production
	Client    *http.Client
}

type openMeteoResponse struct {
	Daily struct {
		TemperatureMax []float64 `json:"temperature_2m_max"`
		TemperatureMin []float64 `json:"temperature_2m_min"`
		Precipitation  []float64 `json:"precipitation_sum"`
		WeatherCode    []int     `json:"weather_code"`
	} `json:"daily"`
}

func (o *OpenMeteo) Today(ctx context.Context) (*Forecast, error) {
	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	base := o.BaseURL
	if base == "" {
		base = "https://api.open-meteo.com"
	}

	params := url.Values{
		"latitude":      {fmt.Sprintf("%f", o.Latitude)},
		"longitude":     {fmt.Sprintf("%f", o.Longitude)},
		"daily":         {"temperature_2m_max,temperature_2m_min,precipitation_sum,weather_code"},
		"timezone":      {o.Timezone},
		"forecast_days": {"1"},
	}
	reqURL := base + "/v1/forecast?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var data openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	d := data.Daily
	if len(d.TemperatureMax) == 0 || len(d.TemperatureMin) == 0 ||
		len(d.Precipitation) == 0 || len(d.WeatherCode) == 0 {
		return nil, fmt.Errorf("empty daily data in response")
	}

	return &Forecast{
		TemperatureMin: d.TemperatureMin[0],
		TemperatureMax: d.TemperatureMax[0],
		Precipitation:  d.Precipitation[0],
		WeatherCode:    d.WeatherCode[0],
	}, nil
}
