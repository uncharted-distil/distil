#!/bin/bash
export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
export ES_ENDPOINT=http://localhost:9200
export PG_STORAGE=true
export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none # debug, error, warn, info, none
export SKIP_INGEST=true
export ROOT_RESOURCE_DIRECTORY=http://localhost:5440
export D3MINPUTDIR_ROOT=`pwd`/datasets
export D3MINPUTDIR=`pwd`/datasets #/LL1_726_TIDY_GPS_carpool_bus_service_rating_prediction #196_autoMpg
export D3MOUTPUTDIR=`pwd`/outputs
export TEMP_STORAGE_ROOT=`pwd`/outputs/temp
export USER_PROBLEM_PATH=`pwd`/outputs/problems
export SOLUTION_SEARCH_MAX_TIME=3

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
