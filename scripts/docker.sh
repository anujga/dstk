#!/bin/bash
set -o errexit


GIT_COMMIT=$(git log -n 1 --format=%h)
IMAGE=localhost:5000/dstk:"$GIT_COMMIT"
docker build -f $1 . -t "$IMAGE"
docker tag "$IMAGE" "localhost:5000/dstk:latest"

docker push "$IMAGE"
docker push "localhost:5000/dstk:latest"