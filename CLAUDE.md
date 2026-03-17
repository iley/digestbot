# Digestbot

Personal Telegram bot that sends a daily digest combining weather and news. Built in Go with a modular segment-based architecture.

## Status

Work in progress. **Milestone 2 (Segment architecture) is complete.** The bot composes a digest from placeholder segments and sends it via Telegram. See the Obsidian note `Digestbot.md` for the full roadmap (M3–M6).

## Architecture

- **Segment-based**: each content source implements the `Segment` interface (`internal/segment/`); the orchestrator (`internal/digest/`) collects and composes output. Segments are responsible for returning valid Telegram HTML (use `segment.EscapeHTML` for dynamic text).
- **All external deps behind Go interfaces**: `WeatherProvider`, `ContentExtractor`, `LLM`, etc.
- **Config**: CLI flags + environment variables only, no config files. Secrets via env vars (`DIGESTBOT_BOT_TOKEN`, `DIGESTBOT_CHAT_ID`).

## Project Layout

```
cmd/digestbot/       — entry point
internal/config/     — CLI flags + env var parsing
internal/segment/    — Segment interface + placeholder implementation
internal/digest/     — orchestrator (Compose)
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
