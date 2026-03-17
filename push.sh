#!/bin/sh
set -e

IMAGE="registry.iley.ru/digestbot:latest"

docker build -t "$IMAGE" .
docker push "$IMAGE"
