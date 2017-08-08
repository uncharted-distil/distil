#!/bin/bash
docker run \
  --user elasticsearch \
  --rm \
  --name distil_dev_es \
  -p 9200:9200 \
  -p 6379:6379 \
  docker.uncharted.software/distil_dev_es:latest
