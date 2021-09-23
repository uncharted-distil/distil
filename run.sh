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

# Docker host IP for the services started by ./run_services.sh
if [[ -z ${DOCKER_HOST} ]]; then
  export SOLUTION_COMPUTE_ENDPOINT=localhost:45042
  export ES_ENDPOINT=http://localhost:9200
else
  export SOLUTION_COMPUTE_ENDPOINT="${DOCKER_HOST}:45042"
  export ES_ENDPOINT="http://${DOCKER_HOST}:9200"
  export PG_HOST="${DOCKER_HOST}"
fi

export SOLUTION_COMPUTE_TRACE=true
export PG_LOG_LEVEL=none # debug, error, warn, info, none
export SOLUTION_SEARCH_MAX_TIME=30000
export SOLUTION_COMPUTE_PULL_MAX=900000
export SOLUTION_COMPUTE_PULL_TIMEOUT=60000
export DATAMART_URL_NYU=https://auctus.vida-nyu.org
export SUMMARY_ENABLED=false # no duke summarization on ingest
export FEATURIZATION_ENABLED=true # featurize remote sensing imagery on ingest
export CLUSTERING_KMEANS=true # 'true' if kmeans should be used for clustering 'false' if we should use hdbscan
export TILE_REQUEST_URL=https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}.png
export INGEST_SAMPLE_ROW_LIMIT=200000
# export MAX_TRAINING_ROWS=500
# export MAX_TEST_ROWS=500

ulimit -n 4096

witch --cmd="make compile && make fmt && go run main.go" --watch="main.go,api/**/*.go" --ignore=""
