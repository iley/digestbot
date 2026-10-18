# Digestbot

Personal Telegram bot that sends a daily digest combining weather and news.

## Digest contents

- **Weather** — today's forecast (temperature range, precipitation, conditions) from [Open-Meteo](https://open-meteo.com/)
- **Irish Times** — top 3 stories from the [Irish Times RSS feed](https://www.irishtimes.com/arc/outboundfeeds/rss/), selected and summarized by an LLM
- **Meduza** — top 3 stories from the [Meduza RSS feed](https://meduza.io/rss/news), selected and summarized in Russian by an LLM

## News processing

Each feed supplies up to 15 candidates. A `news.Ranker` selects the three most important stories (or all available stories for smaller feeds), then a `news.Summarizer` writes one sentence per selected story. By default both use the configured LLM through separate calls. Set `--news-ranker=jev` to rank with Jev while keeping LLM summarization. Meduza summaries are in Russian.

Selections are validated for count, bounds, and uniqueness; summaries must match the selected articles in count and order. With neither component configured, the news segment renders up to five raw headlines.

### Jev ranking

The Jev ranker uses [TypeSafe’s Score API](https://docs.typesafe.ai/primitives/score): one question per candidate in a single request. It scores public significance on a five-level rubric, sorts descending, and preserves feed order for ties. The rubric is in `internal/news/jev.go`. It does not deduplicate stories about the same event.

With the usual Telegram, feed, and LLM configuration already set:

```sh
export DIGESTBOT_NEWS_RANKER=jev
export DIGESTBOT_JEV_API_KEY=your-typesafe-api-key
./digestbot
```

The LLM API key is still required for summaries. Requests have a 30-second timeout; API errors or missing/invalid scores fail the news segment rather than silently switching providers. Set `DIGESTBOT_NEWS_RANKER=llm` to switch back. `jev-latest` follows TypeSafe’s latest model; use `DIGESTBOT_JEV_MODEL` to choose a specific supported version.

### Verifying the deployed ranker

Logs go to stderr and appear in your container/runtime logs (for Docker, `docker logs <container>`). Startup logs show the configured provider and enabled segments. A successful ranking logs the provider, candidate and selected counts, and elapsed duration. Successful LLM summarization logs its article count and elapsed duration. Jev also logs the requested model and the model version returned by the API (`unknown` if omitted). For example:

```text
news ranking configured provider="jev" segments=["weather" "irishtimes" "meduza"]
news ranking succeeded provider=jev requested_model="jev-latest" model="jev-1.13.0" candidates=15 selected=3 duration=420ms
news summarization succeeded provider=llm articles=3 duration=1.2s
```

These are illustrative values. The success line confirms a completed, validated ranking call; the startup line only confirms configuration. Empty feeds do not produce ranking success logs. Ranking success does not confirm summarization or Telegram delivery; segment failures are logged separately. Ranking logs do not include API keys or article content.

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
| News ranker | `DIGESTBOT_NEWS_RANKER` | `--news-ranker` | no | `llm` (`llm` or `jev`) |
| Jev API key | `DIGESTBOT_JEV_API_KEY` | `--jev-api-key` | when Jev ranks news | — |
| Jev model | `DIGESTBOT_JEV_MODEL` | `--jev-model` | no | `jev-latest` |
| Jev API base URL | `DIGESTBOT_JEV_BASE_URL` | `--jev-base-url` | no | `https://api.typesafe.ai/v1` |
| Timezone | `DIGESTBOT_TIMEZONE` | `--timezone` | no | `Europe/Dublin` |
| Segments | `DIGESTBOT_SEGMENTS` | `--segments` | no | `weather,irishtimes,meduza` |

Latitude and longitude are used to fetch weather from [Open-Meteo](https://open-meteo.com/). Timezone must be a valid [tz database name](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) (e.g. `Europe/Dublin`, `America/New_York`).

Segments controls which digest sections are included and in what order. Available segments: `weather`, `irishtimes`, `meduza`.

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
