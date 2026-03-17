package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/iley/digestbot/internal/config"
	"github.com/iley/digestbot/internal/digest"
	"github.com/iley/digestbot/internal/segment"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	segments := []segment.Segment{
		&segment.Placeholder{Title: "Weather", Body: "Dublin: 12°C, partly cloudy."},
		&segment.Placeholder{Title: "Irish Times", Body: "Top story: placeholder headline."},
		&segment.Placeholder{Title: "Meduza", Body: "Главное: заглушка заголовка."},
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
