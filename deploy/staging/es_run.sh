#!/bin/sh

docker run \
  --label=com.centurylinklabs.watchtower.stop-signal=SIGKILL \
  --user elasticsearch \
  --rm \
  --name distil_dev_es \
  --net=distil_nw \
  docker.uncharted.software/distil_dev_es
