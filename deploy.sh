#!/bin/bash
export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export SOLUTION_DATA_DIR=/datasets
export PG_STORAGE=true
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none

yarn build
make build
sudo --preserve-env ./distil
