#!/bin/bash

# builds distil docker image
pushd .
cd ..
make build_static
yarn build
popd
docker build -t docker.uncharted.software/distil:latest ..
