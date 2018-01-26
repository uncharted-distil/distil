#!/bin/bash
export PIPELINE_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export PIPELINE_DATA_DIR=/datasets
export PG_STORAGE=true
export PIPELINE_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none

yarn build
make build
sudo --preserve-env ./distil
