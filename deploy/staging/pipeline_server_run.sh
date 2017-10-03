#!/bin/sh

docker run \
  --rm \
  --name distil-pipeline-server \
  --net=distil_nw \
  --label=com.centurylinklabs.watchtower.stop-signal=SIGKILL \
  -p 9500:9500 \
  -v `pwd`/datasets:`pwd`/datasets \
  -e PIPELINE_SERVER_RESULT_DIR=`pwd`/datasets \
  docker.uncharted.software/distil-pipeline-server
