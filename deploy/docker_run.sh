#!/bin/bash
docker run \
   --name distil \
    --rm \
    -d \
    -p 8080:8080 \
    -e PIPELINE_COMPUTE_ENDPOINT=localhost:45042 \
    -e ES_ENDPOINT=http://localhost:9200 \
    -e PIPELINE_DATA_DIR=`pwd`/datasets \
    -e PG_STORAGE=true \
    -e PIPELINE_COMPUTE_TRACE=true \
    -e PG_LOG_LEVEL=none \
    --entrypoint ta3_search \
    docker.uncharted.software/distil:latest
 
