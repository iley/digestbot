# Digestbot

Personal Telegram bot that sends a daily digest combining weather and news. Built in Go with a modular segment-based architecture.

## Status

Work in progress. **Milestone 6 (Meduza integration) is complete.** The bot fetches Irish Times and Meduza news, uses an LLM to pick the 3 most important from each source and summarize them (Meduza in Russian), then sends the digest via Telegram. See the Obsidian note `Digestbot.md` for the full roadmap.

## Architecture

- **Segment-based**: each content source implements the `Segment` interface (`internal/segment/`); the orchestrator (`internal/digest/`) collects and composes output. Segments are registered by name in `cmd/digestbot/main.go` and selected via the `--segments` flag (default: `weather,irishtimes,meduza`). Segments are responsible for returning valid Telegram HTML (use `segment.EscapeHTML` for dynamic text).
- **All external deps behind Go interfaces**: `WeatherProvider`, `ContentExtractor`, `LLM`, etc.
- **Config**: CLI flags + environment variables only, no config files. Secrets via env vars (`DIGESTBOT_BOT_TOKEN`, `DIGESTBOT_CHAT_ID`, `DIGESTBOT_OPENAI_API_KEY`).

## Project Layout

```
cmd/digestbot/       — entry point
internal/config/     — CLI flags + env var parsing
internal/segment/    — Segment interface + implementations
internal/digest/     — orchestrator (Compose)
internal/llm/        — LLM interface + OpenAI implementation
internal/news/       — news feed fetching (Article, FeedFetcher, RSSFetcher)
internal/weather/    — weather data fetching
```

## Build & Run

```sh
# Build
go build ./cmd/digestbot

# Run (requires env vars or flags)
DIGESTBOT_BOT_TOKEN=... DIGESTBOT_CHAT_ID=... go run ./cmd/digestbot

# Docker
docker build -t digestbot .
```

## After Making Changes

Always run before committing:

```sh
go vet ./...
go fmt ./...
```
