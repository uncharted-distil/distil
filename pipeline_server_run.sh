mkdir -p datasets
docker run \
  --rm \
  --name distil_pipeline_server \
  -p 50051:50051 \
  -v `pwd`/datasets:`pwd`/datasets \
  -e PIPELINE_SERVER_RESULT_DIR=`pwd`/datasets \
  docker.uncharted.software/distil-pipeline-server
