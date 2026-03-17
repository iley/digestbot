FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /digestbot ./cmd/digestbot

FROM alpine:3.21
COPY --from=builder /digestbot /digestbot
ENTRYPOINT ["/digestbot"]
