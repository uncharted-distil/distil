#!/bin/bash

source ./config.sh

# builds distil docker image
pushd .
cd ..
make build
yarn build
popd
docker build -t $DOCKER_REPO/$DOCKER_IMAGE_NAME:latest ..
