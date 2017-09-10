#!/bin/bash

export ES_ENDPOINT=http://localhost:9200
export PIPELINE_DATA_DIR=`pwd`/datasets
export PG_STORAGE=false
witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
