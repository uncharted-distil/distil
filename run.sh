#!/bin/bash
export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=debug # debug, error, warn, info, none
export SKIP_INGEST=true
export D3MINPUTDIR=~/data/d3m
export D3MOUTPUTDIR=~/temp/outputs
export USER_PROBLEM_PATH=`pwd`/outputs/problems
export SOLUTION_SEARCH_MAX_TIME=3
export SOLUTION_COMPUTE_PULL_MAX=900
export SOLUTION_COMPUTE_TIMEOUT=600
export USE_TA2_RUNNER=true #false
export DATAMART_IMPORT_FOLDER=`pwd`/datamart
export TEMP_STORAGE_ROOT=~/temp/outputs
export DATA_FOLDER_PATH=~/data/d3m

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
