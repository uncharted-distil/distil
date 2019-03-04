#!/bin/bash

source ./config.sh

# builds distil docker image
pushd .
cd ..
make build_static
yarn build
popd
docker build -t $DOCKER_REPO/$IMAGE_NAME:latest ..
