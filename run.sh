#!/bin/bash

if [[ -z "${D3MINPUTDIR}" ]]; then
  export D3MINPUTDIR="$PWD/datasets"
fi

if [[ -z "${D3MOUTPUTDIR}" ]]; then
  export D3MOUTPUTDIR="$PWD/outputs"
fi

if [[ -z "${DATAMART_IMPORT_FOLDER}" ]]; then
  export DATAMART_IMPORT_FOLDER="$PWD/datamart"
fi

export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none # debug, error, warn, info, none
export SKIP_INGEST=true
export SOLUTION_SEARCH_MAX_TIME=30000
export SOLUTION_COMPUTE_PULL_MAX=900000
export SOLUTION_COMPUTE_PULL_TIMEOUT=60000
export DATAMART_URL_NYU=https://auctus.vida-nyu.org
# export MAX_TRAINING_ROWS=500
# export MAX_TEST_ROWS=500

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
