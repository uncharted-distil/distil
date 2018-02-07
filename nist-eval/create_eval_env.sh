# takes a single argument - the path of the dateset to act as the ingest targeti
#
# Requires Jinja2 library, jinja2-cli:
#
# pip install jinja2 jinja2-cli

echo "Removing previous environment"
sudo chgrp -R $USER /tmp/d3m
sudo chown -R $USER /tmp/d3m
chmod -R u+w /tmp/d3m
rm -rf /tmp/d3m

echo "Creating directory hieararchy"
mkdir -p /tmp/d3m/executables
mkdir /tmp/d3m/config
mkdir /tmp/d3m/dataset
mkdir /tmp/d3m/logs
mkdir /tmp/d3m/temp_storage

echo "Copying $1m"
cp -r $1 /tmp/d3m/dataset

echo "Generating config"
jinja2 ./config.json -Ddataset_name=`basename $1` > /tmp/d3m/config/config.json


chmod -R 555 /tmp/d3m
chmod 777 /tmp/d3m/temp_storage

echo "Setting JSON_CONFIG_PATH" 
export JSON_CONFIG_PATH=/tmp/d3m/config/config.json
