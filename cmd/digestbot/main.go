package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/iley/digestbot/internal/config"
	"github.com/iley/digestbot/internal/digest"
	"github.com/iley/digestbot/internal/llm"
	"github.com/iley/digestbot/internal/news"
	"github.com/iley/digestbot/internal/segment"
	"github.com/iley/digestbot/internal/weather"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	weatherProvider := &weather.OpenMeteo{
		Latitude:  cfg.Latitude,
		Longitude: cfg.Longitude,
		Timezone:  cfg.Timezone,
	}

	llmClient := &llm.OpenAI{
		APIKey:  cfg.OpenAIAPIKey,
		Model:   cfg.LLMModel,
		BaseURL: cfg.LLMBaseURL,
	}
	irishtimesFetcher := &news.RSSFetcher{FeedURL: "https://www.irishtimes.com/arc/outboundfeeds/rss/"}
	meduzaFetcher := &news.RSSFetcher{FeedURL: "https://meduza.io/rss/news"}

	irishtimes := &segment.News{Title: "Irish Times", Fetcher: irishtimesFetcher, LLM: llmClient}
	meduza := &segment.News{Title: "Meduza", Fetcher: meduzaFetcher, LLM: llmClient, Language: "ru"}

	available := map[string]digest.NamedSegment{
		"weather":    {Name: "Weather", Segment: &segment.Weather{Provider: weatherProvider}},
		"irishtimes": {Name: irishtimes.Title, Segment: irishtimes},
		"meduza":     {Name: meduza.Title, Segment: meduza},
	}

	segments := make([]digest.NamedSegment, 0, len(cfg.Segments))
	for _, name := range cfg.Segments {
		seg, ok := available[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "error: unknown segment %q\n", name)
			os.Exit(1)
		}
		segments = append(segments, seg)
	}

	ctx := context.Background()
	text := digest.Compose(ctx, segments)

	b, err := bot.New(cfg.BotToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating bot: %v\n", err)
		os.Exit(1)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    cfg.ChatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Message sent successfully.")
}
