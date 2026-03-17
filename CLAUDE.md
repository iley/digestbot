# Digestbot

Personal Telegram bot that sends a daily digest combining weather and news. Built in Go with a modular segment-based architecture.

## Status

Work in progress. **Milestone 1 (Telegram bot boilerplate) is complete.** The bot currently sends a hardcoded test message and exits. See the Obsidian note `Digestbot.md` for the full roadmap (M2–M6).

## Architecture

- **Segment-based**: each content source will implement a `Segment` interface; an orchestrator collects and composes output.
- **All external deps behind Go interfaces**: `WeatherProvider`, `ContentExtractor`, `LLM`, etc.
- **Config**: CLI flags + environment variables only, no config files. Secrets via env vars (`DIGESTBOT_BOT_TOKEN`, `DIGESTBOT_CHAT_ID`).

## Project Layout

```
cmd/digestbot/   — entry point
internal/        — internal packages (config, future segments, etc.)
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
