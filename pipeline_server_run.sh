docker run \
  --rm \
  --name distil_pipeline_server \
  -p 9500:9500 \
  -v `pwd`/datasets:`pwd`/datasets \
  -e PIPELINE_SERVER_RESULT_DIR=`pwd`/datasets \
  docker.uncharted.software/distil-pipeline-server
