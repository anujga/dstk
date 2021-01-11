#!/bin/bash
set -o errexit

GIT_COMMIT=$(git log -n 1 --format=%h)
IMAGE=localhost:5000/dstk:"$GIT_COMMIT"
docker build -f $1 . -t "$IMAGE"
docker tag "$IMAGE" "localhost:5000/dstk:latest"
docker push "localhost:5000/dstk:latest"

#docker tag "$IMAGE" "004245605591.dkr.ecr.ap-south-1.amazonaws.com/dstk:latest"
#docker push "004245605591.dkr.ecr.ap-south-1.amazonaws.com/dstk:latest"
