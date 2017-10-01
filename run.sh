#!/bin/bash
export PIPELINE_COMPUTE_ENDPOINT=localhost:50051
export ES_ENDPOINT=http://localhost:9200
export PIPELINE_DATA_DIR=`pwd`/datasets
export PG_STORAGE=true
export PIPELINE_COMPUTE_TRACE=true
witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
