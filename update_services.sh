#!/bin/bash

if [[ -z "${D3MINPUTDIR}" ]]; then
  export D3MINPUTDIR="$PWD/datasets"
fi

if [[ -z "${D3MOUTPUTDIR}" ]]; then
    export D3MOUTPUTDIR="$PWD/outputs"
fi

if [[ -z "${D3MSTATICDIR}" ]]; then
    export D3MSTATICDIR="$PWD/static_resources"
fi

if [[ -z "${DATAMART_IMPORT_FOLDER}" ]]; then
  export DATAMART_IMPORT_FOLDER="$PWD/datamart"
fi

 docker-compose pull
