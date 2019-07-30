#!/bin/bash

if [[ -z "${D3MINPUTDIR}" ]]; then
  export D3MINPUTDIR="$PWD/datasets"
fi

if [[ -z "${D3MOUTPUTDIR}" ]]; then
  export D3MOUTPUTDIR="$PWD/outputs"
fi

export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none # debug, error, warn, info, none
export SKIP_INGEST=true
export SOLUTION_SEARCH_MAX_TIME=3
export SOLUTION_COMPUTE_PULL_MAX=900
export SOLUTION_COMPUTE_TIMEOUT=600
export USE_TA2_RUNNER=true #false

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
