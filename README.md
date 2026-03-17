# Digestbot

Personal Telegram bot that sends a daily digest combining weather and news.

## Digest contents

- **Weather** — today's forecast (temperature range, precipitation, conditions) from [Open-Meteo](https://open-meteo.com/)
- **Irish Times** — top 3 stories from the [Irish Times RSS feed](https://www.irishtimes.com/arc/outboundfeeds/rss/), selected and summarized by an LLM

## Prerequisites

- Go 1.26+
- A Telegram bot token (from [@BotFather](https://t.me/BotFather))
- The chat ID where the bot should send messages (send a message to your bot, then visit `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates` — the chat ID is in `result[].message.chat.id`)
- An OpenAI API key (from [platform.openai.com](https://platform.openai.com/))

## Configuration

All configuration is done via environment variables or CLI flags. Flags take precedence over environment variables.

| Setting | Env var | Flag | Required | Default |
|---------|---------|------|----------|---------|
| Telegram bot token | `DIGESTBOT_BOT_TOKEN` | `--bot-token` | yes | — |
| Telegram chat ID | `DIGESTBOT_CHAT_ID` | `--chat-id` | yes | — |
| Latitude | `DIGESTBOT_LATITUDE` | `--latitude` | yes | — |
| Longitude | `DIGESTBOT_LONGITUDE` | `--longitude` | yes | — |
| OpenAI API key | `DIGESTBOT_OPENAI_API_KEY` | `--openai-api-key` | yes | — |
| Timezone | `DIGESTBOT_TIMEZONE` | `--timezone` | no | `Europe/Dublin` |

Latitude and longitude are used to fetch weather from [Open-Meteo](https://open-meteo.com/). Timezone must be a valid [tz database name](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) (e.g. `Europe/Dublin`, `America/New_York`).

## Running

### From source

```sh
go build ./cmd/digestbot

DIGESTBOT_BOT_TOKEN=123:ABC \
DIGESTBOT_CHAT_ID=-100123456 \
DIGESTBOT_LATITUDE=53.35 \
DIGESTBOT_LONGITUDE=-6.26 \
DIGESTBOT_OPENAI_API_KEY=sk-... \
./digestbot
```

Or with flags:

```sh
./digestbot --bot-token 123:ABC --chat-id -100123456 --latitude 53.35 --longitude -6.26 --openai-api-key sk-...
```

### With `go run`

```sh
DIGESTBOT_BOT_TOKEN=123:ABC \
DIGESTBOT_CHAT_ID=-100123456 \
DIGESTBOT_LATITUDE=53.35 \
DIGESTBOT_LONGITUDE=-6.26 \
DIGESTBOT_OPENAI_API_KEY=sk-... \
go run ./cmd/digestbot
```

### Docker

```sh
docker build -t digestbot .

docker run --rm \
  -e DIGESTBOT_BOT_TOKEN=123:ABC \
  -e DIGESTBOT_CHAT_ID=-100123456 \
  -e DIGESTBOT_LATITUDE=53.35 \
  -e DIGESTBOT_LONGITUDE=-6.26 \
  -e DIGESTBOT_OPENAI_API_KEY=sk-... \
  digestbot
```

## Development

```sh
go vet ./...
go fmt ./...
go test ./...
```
