#!/bin/bash

# datasets we want
declare -a datasets=("185_baseball" "LL0_acled" "22_handgeometry")

mkdir temp_datasets
cd ./temp_datasets

echo "Cloning datasets repo"

git lfs clone git@gitlab.datadrivendiscovery.org:d3m/datasets.git -X "*"

cd ./datasets

for dataset in "${datasets[@]}"
do
   echo "Pulling LFS files for $dataset"

   git lfs pull -I seed_datasets_current/$dataset/

   echo "Copying $dataset to datasets dir"

   mv seed_datasets_current/$dataset ../../datasets/$dataset
done

echo "Removing temporary files"

cd ../../
rm -rf ./temp_datasets
