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

	llmClient := &llm.OpenAI{APIKey: cfg.OpenAIAPIKey}
	irishtimesFetcher := &news.RSSFetcher{FeedURL: "https://www.irishtimes.com/arc/outboundfeeds/rss/"}
	meduzaFetcher := &news.RSSFetcher{FeedURL: "https://meduza.io/rss/news"}

	segments := []segment.Segment{
		&segment.Weather{Provider: weatherProvider},
		&segment.News{Title: "Irish Times", Fetcher: irishtimesFetcher, LLM: llmClient},
		&segment.News{Title: "Meduza", Fetcher: meduzaFetcher, LLM: llmClient, Language: "ru"},
	}

	ctx := context.Background()
	text, err := digest.Compose(ctx, segments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error composing digest: %v\n", err)
		os.Exit(1)
	}

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
