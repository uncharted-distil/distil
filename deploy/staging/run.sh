#!/bin/bash
docker run \
    --name distil \
    --network distil_nw \
    --label=com.centurylinklabs.watchtower.stop-signal=SIGKILL \
    --rm \
    -p 8083:8080 \
    -e PG_STORAGE=true \
    -e PG_HOST=distil_dev_postgres \
    -e ES_ENDPOINT=http://distil_dev_es:9200 \
    -e REDIS_ENDPOINT=distil_dev_es:6379 \
    -e PIPELINE_COMPUTE_ENDPOINT=distil-pipeline-server:9500 \
    -e PIPELINE_DATA_DIR=`pwd`/datasets \
    -v `pwd`/datasets:`pwd`/datasets \
    docker.uncharted.software/distil
