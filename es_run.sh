#!/bin/bash
docker run \
  --user elasticsearch \
  --rm \
  --name distil_dev_es \
  -p 9200:9200 \
  docker.uncharted.software/distil_dev_es:latest
