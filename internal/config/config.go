package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken     string
	ChatID       int64
	Latitude     float64
	Longitude    float64
	Timezone     string
	OpenAIAPIKey string
	Segments     []string
}

func Parse() (*Config, error) {
	botTokenDefault := os.Getenv("DIGESTBOT_BOT_TOKEN")
	chatIDDefault := int64(0)
	if s := os.Getenv("DIGESTBOT_CHAT_ID"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid DIGESTBOT_CHAT_ID: %w", err)
		}
		chatIDDefault = v
	}

	latDefault, lonDefault := 0.0, 0.0
	latSet, lonSet := false, false
	if s := os.Getenv("DIGESTBOT_LATITUDE"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid DIGESTBOT_LATITUDE: %w", err)
		}
		latDefault = v
		latSet = true
	}
	if s := os.Getenv("DIGESTBOT_LONGITUDE"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid DIGESTBOT_LONGITUDE: %w", err)
		}
		lonDefault = v
		lonSet = true
	}
	openAIKeyDefault := os.Getenv("DIGESTBOT_OPENAI_API_KEY")
	tzDefault := os.Getenv("DIGESTBOT_TIMEZONE")
	if tzDefault == "" {
		tzDefault = "Europe/Dublin"
	}
	segmentsDefault := os.Getenv("DIGESTBOT_SEGMENTS")
	if segmentsDefault == "" {
		segmentsDefault = "weather,irishtimes,meduza"
	}

	fs := flag.NewFlagSet("digestbot", flag.ContinueOnError)
	botToken := fs.String("bot-token", botTokenDefault, "Telegram bot token (env: DIGESTBOT_BOT_TOKEN)")
	chatID := fs.Int64("chat-id", chatIDDefault, "Telegram chat ID (env: DIGESTBOT_CHAT_ID)")
	latitude := fs.Float64("latitude", latDefault, "Latitude for weather (env: DIGESTBOT_LATITUDE)")
	longitude := fs.Float64("longitude", lonDefault, "Longitude for weather (env: DIGESTBOT_LONGITUDE)")
	timezone := fs.String("timezone", tzDefault, "Timezone for weather (env: DIGESTBOT_TIMEZONE)")
	openAIKey := fs.String("openai-api-key", openAIKeyDefault, "OpenAI API key (env: DIGESTBOT_OPENAI_API_KEY)")
	segmentsFlag := fs.String("segments", segmentsDefault, "Comma-separated list of segments (env: DIGESTBOT_SEGMENTS)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// Parse and deduplicate segments first, so we can validate only the
	// config that the selected segments actually need.
	var segments []string
	seen := make(map[string]bool)
	for _, s := range strings.Split(*segmentsFlag, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if seen[s] {
			return nil, fmt.Errorf("duplicate segment %q", s)
		}
		seen[s] = true
		segments = append(segments, s)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("at least one segment is required (--segments or DIGESTBOT_SEGMENTS)")
	}

	// TODO: Unify this segment requirement metadata with the segment builder
	// registry in cmd/digestbot/main.go so new segments don't require updates
	// in two places.
	needsWeather := seen["weather"]
	needsLLM := seen["irishtimes"] || seen["meduza"]

	if *botToken == "" {
		return nil, fmt.Errorf("bot token is required (--bot-token or DIGESTBOT_BOT_TOKEN)")
	}
	if *chatID == 0 {
		return nil, fmt.Errorf("chat ID is required (--chat-id or DIGESTBOT_CHAT_ID)")
	}
	if needsLLM && *openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required (--openai-api-key or DIGESTBOT_OPENAI_API_KEY)")
	}

	if needsWeather {
		// Check if latitude/longitude were set via flags.
		flagLatSet, flagLonSet := false, false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "latitude" {
				flagLatSet = true
			}
			if f.Name == "longitude" {
				flagLonSet = true
			}
		})
		if !latSet && !flagLatSet {
			return nil, fmt.Errorf("latitude is required (--latitude or DIGESTBOT_LATITUDE)")
		}
		if !lonSet && !flagLonSet {
			return nil, fmt.Errorf("longitude is required (--longitude or DIGESTBOT_LONGITUDE)")
		}
	}

	return &Config{
		BotToken:     *botToken,
		ChatID:       *chatID,
		Latitude:     *latitude,
		Longitude:    *longitude,
		Timezone:     *timezone,
		OpenAIAPIKey: *openAIKey,
		Segments:     segments,
	}, nil
}
