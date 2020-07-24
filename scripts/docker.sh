#!/bin/bash

GIT_COMMIT=$(git log -n 1 --format=%h)
IMAGE=localhost:5000/dstk:"$GIT_COMMIT"
docker build -f deploy/Dockerfile . -t "$IMAGE" && \
docker push "$IMAGE"
