#!/bin/bash
export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none # debug, error, warn, info, none
export SKIP_INGEST=true
export ROOT_RESOURCE_DIRECTORY=http://localhost:5440
export D3MINPUTDIR_ROOT=`pwd`/datasets
export D3MINPUTDIR=`pwd`/datasets
export USER_PROBLEM_PATH=`pwd`/outputs/problems
export SOLUTION_SEARCH_MAX_TIME=3
export SOLUTION_COMPUTE_PULL_MAX=900
export USE_TA2_RUNNER=true #false
export DATAMART_IMPORT_FOLDER=`pwd`/datamart

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
