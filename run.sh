#!/bin/bash

export ES_ENDPOINT=http://10.64.16.120:9200
witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
