#!/bin/sh
set -e

IMAGE="registry.iley.ru/digestbot:latest"

docker build --platform linux/amd64 -t "$IMAGE" .
docker push "$IMAGE"
