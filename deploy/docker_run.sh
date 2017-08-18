#!/bin/bash
docker run \
    --name distil \
    --rm \
    -p 9000:8080 \
    -e ES_ENDPOINT=http://10.64.16.120:9200 \
    -e MARVIN_ENDPOINT=http://d3m-dev.dyndns.org:80/es \
    docker.uncharted.software/distil:0.1
