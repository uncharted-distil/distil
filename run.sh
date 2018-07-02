#!/bin/bash
export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export D3MINPUTDIR=`pwd`/datasets
export PG_STORAGE=true
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none # debug, error, warn, info, none
export SKIP_INGEST=true
export ROOT_RESOURCE_DIRECTORY=http://localhost:5440
export TEMP_STORAGE_ROOT=datasets
export D3MOUTPUTDIR=datasets
export USER_PROBLEM_PATH=output

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
