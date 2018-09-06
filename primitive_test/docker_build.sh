#!/bin/bash

# copy local primitive source
rm -rf ./common-primitives
cp -r $PRIMITIVE_SRC_DIR/common-primitives .

# build the container
docker build --no-cache -t primitive_test .

