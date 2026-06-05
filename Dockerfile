FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /digestbot ./cmd/digestbot

FROM alpine:3.21
# CA certs for HTTPS calls (Telegram, Open-Meteo, the LiteLLM proxy).
RUN apk add --no-cache ca-certificates
COPY --from=builder /digestbot /digestbot
ENTRYPOINT ["/digestbot"]
