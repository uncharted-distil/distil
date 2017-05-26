#!/bin/bash
docker run \
    --name distil-server \
    --rm \
    -p 9000:8080 \
    -e DATASET_ENDPOINT=http://10.64.16.120:9200 \
    -e MARVIN_ENDPOINT=http://d3m-dev.dyndns.org:80/es \
    docker.uncharted.software/distil-server:0.1
