package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const cannedResponse = `{
  "daily": {
    "temperature_2m_max": [12.3],
    "temperature_2m_min": [5.1],
    "precipitation_sum": [0.5],
    "weather_code": [2]
  }
}`

func TestOpenMeteoToday(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("latitude") == "" || q.Get("longitude") == "" {
			t.Error("missing latitude/longitude query params")
		}
		if q.Get("timezone") == "" {
			t.Error("missing timezone query param")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cannedResponse))
	}))
	defer srv.Close()

	client := &OpenMeteo{
		Latitude:  53.35,
		Longitude: -6.26,
		Timezone:  "Europe/Dublin",
		BaseURL:   srv.URL,
	}

	forecast, err := client.Today(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if forecast.TemperatureMax != 12.3 {
		t.Errorf("TemperatureMax = %v, want 12.3", forecast.TemperatureMax)
	}
	if forecast.TemperatureMin != 5.1 {
		t.Errorf("TemperatureMin = %v, want 5.1", forecast.TemperatureMin)
	}
	if forecast.Precipitation != 0.5 {
		t.Errorf("Precipitation = %v, want 0.5", forecast.Precipitation)
	}
	if forecast.WeatherCode != 2 {
		t.Errorf("WeatherCode = %v, want 2", forecast.WeatherCode)
	}
}

func TestOpenMeteoEmptyDaily(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"daily": {"temperature_2m_max": [], "temperature_2m_min": [], "precipitation_sum": [], "weather_code": []}}`))
	}))
	defer srv.Close()

	client := &OpenMeteo{
		Latitude:  53.35,
		Longitude: -6.26,
		Timezone:  "Europe/Dublin",
		BaseURL:   srv.URL,
	}

	_, err := client.Today(context.Background())
	if err == nil {
		t.Fatal("expected error for empty daily data")
	}
}

func TestOpenMeteoServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := &OpenMeteo{
		Latitude:  53.35,
		Longitude: -6.26,
		Timezone:  "Europe/Dublin",
		BaseURL:   srv.URL,
	}

	_, err := client.Today(context.Background())
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}
