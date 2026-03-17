package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	BotToken string
	ChatID   int64
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

	fs := flag.NewFlagSet("digestbot", flag.ContinueOnError)
	botToken := fs.String("bot-token", botTokenDefault, "Telegram bot token (env: DIGESTBOT_BOT_TOKEN)")
	chatID := fs.Int64("chat-id", chatIDDefault, "Telegram chat ID (env: DIGESTBOT_CHAT_ID)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if *botToken == "" {
		return nil, fmt.Errorf("bot token is required (--bot-token or DIGESTBOT_BOT_TOKEN)")
	}
	if *chatID == 0 {
		return nil, fmt.Errorf("chat ID is required (--chat-id or DIGESTBOT_CHAT_ID)")
	}

	return &Config{
		BotToken: *botToken,
		ChatID:   *chatID,
	}, nil
}
