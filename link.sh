#!/bin/bash
rm -rf ./vendor/github.com/unchartedsoftware/distil-compute
rm -rf ./vendor/github.com/unchartedsoftware/distil-ingest

ln -s $GOPATH/src/github.com/unchartedsoftware/distil-compute ./vendor/github.com/unchartedsoftware/distil-compute
ln -s $GOPATH/src/github.com/unchartedsoftware/distil-ingest ./vendor/github.com/unchartedsoftware/distil-ingest
