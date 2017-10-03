#!/bin/bash

docker run \
  --net=distil_nw \
  --name redis \
  -p 6379:6379 \
  redis
