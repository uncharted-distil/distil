#!/bin/bash

source ./config.sh

docker run \
   --name distil \
    --rm \
    -p 8080:8080 \
    -e SOLUTION_COMPUTE_ENDPOINT=localhost:45042 \
    -e ES_ENDPOINT=http://localhost:9200 \
    -e D3MINPUTDIR=$D3MINPUTDIR \
    -e SOLUTION_COMPUTE_TRACE=true \
    -e PG_LOG_LEVEL=none \
    -v $D3MINPUTDIR:$D3MINPUTDIR \
    $DOCKER_REPO/$DOCKER_IMAGE_NAME:latest
