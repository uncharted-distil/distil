#!/bin/bash
docker run \
    --name distil \
    --rm \
    -p 9000:8080 \
    -e ES_ENDPOINT=http://:localhost:9200 \
    docker.uncharted.software/distil:latest \
    ta3_search
