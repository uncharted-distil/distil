#!/bin/bash
killbg() {
        for p in "${pids[@]}" ; do
                kill "$p";
        done
}
trap killbg EXIT
docker restart distil_distil-auto-ml_1
docker restart distil_postgres_1
docker restart distil_elastic_1
pids=()
./run_services.sh & 
pids+=($!)
yarn watch  
