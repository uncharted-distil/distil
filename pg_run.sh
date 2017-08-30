#!/bin/bash
docker run \
  --user postgres \
  --rm \
  -p 5432:5432 \
  --name distil_dev_postgres \
  docker.uncharted.software/distil_dev_postgres:latest \
  -d postgres
