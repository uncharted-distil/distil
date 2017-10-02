#!/bin/bash
docker run \
  --label=com.centurylinklabs.watchtower.stop-signal=SIGKILL \
  --user postgres \
  --rm \
  --name distil_dev_postgres \
  --net=distil_nw \
  -p 5432:5432 \
  docker.uncharted.software/distil_dev_postgres:latest \
  -d postgres
