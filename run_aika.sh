docker run \
  --rm \
  --shm-size=1g \
  --detach \
  --name aika \
  --publish 50051:50051 \
  --volume `pwd`/datasets:/datasets \
  registry.datadrivendiscovery.org/berkeley/aika
