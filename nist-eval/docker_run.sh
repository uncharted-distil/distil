#!/bin/bash
docker run \
    --name distil \
    --network nisteval_default \
    --rm \
    -p 8080:8080 \
    -e SOLUTION_COMPUTE_ENDPOINT=pipeline_server:45042 \
    -e ES_ENDPOINT=http://elastic:9200 \
    -e PG_HOST=postgres \
    -e D3MINPUTDIR=/tmp/d3m/temp_storage \
    -e SOLUTION_COMPUTE_TRACE=true \
    -e PG_LOG_LEVEL=none \
    -e SKIP_INGEST=false \
    -e JSON_CONFIG_PATH=/tmp/d3m/config/config.json \
    -e CLASSIFICATION_ENDPOINT=http://nk_classification_rest:5000 \
    -e RANKING_ENDPOINT=http://nk_ranking_rest:5002 \
    -v /tmp/d3m:/tmp/d3m \
    --entrypoint ta3_search \
    docker.uncharted.software/distil:latest
