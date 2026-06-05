package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const maxAttempts = 3

type OpenMeteo struct {
	Latitude  float64
	Longitude float64
	Timezone  string
	BaseURL   string        // override for testing; empty means production
	Client    *http.Client  // optional, for testing
	Backoff   time.Duration // base delay between retries; zero means 1s
}

type openMeteoResponse struct {
	Daily struct {
		TemperatureMax []float64 `json:"temperature_2m_max"`
		TemperatureMin []float64 `json:"temperature_2m_min"`
		Precipitation  []float64 `json:"precipitation_sum"`
		WeatherCode    []int     `json:"weather_code"`
	} `json:"daily"`
}

// Today fetches today's forecast. It retries a few times on transient failures
// (network errors and 5xx responses), which open-meteo returns intermittently.
func (o *OpenMeteo) Today(ctx context.Context) (*Forecast, error) {
	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	backoff := o.Backoff
	if backoff == 0 {
		backoff = time.Second
	}

	reqURL := o.forecastURL()

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		forecast, err := fetchForecast(ctx, client, reqURL)
		if err == nil {
			return forecast, nil
		}

		var te *transientError
		if !errors.As(err, &te) {
			return nil, err
		}
		lastErr = err

		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * backoff):
			}
		}
	}
	return nil, fmt.Errorf("after %d attempts: %w", maxAttempts, lastErr)
}

func fetchForecast(ctx context.Context, client *http.Client, reqURL string) (*Forecast, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &transientError{fmt.Errorf("fetching weather: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status %d", resp.StatusCode)
		if resp.StatusCode >= 500 {
			return nil, &transientError{err}
		}
		return nil, err
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

func (o *OpenMeteo) forecastURL() string {
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
	return base + "/v1/forecast?" + params.Encode()
}

// transientError marks an error worth retrying.
type transientError struct{ err error }

func (e *transientError) Error() string { return e.err.Error() }
func (e *transientError) Unwrap() error { return e.err }
