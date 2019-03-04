#!/bin/bash
rm -rf ./vendor/github.com/uncharted-distil/distil-compute
rm -rf ./vendor/github.com/uncharted-distil/distil-ingest

ln -s $GOPATH/src/github.com/uncharted-distil/distil-compute ./vendor/github.com/uncharted-distil/distil-compute
ln -s $GOPATH/src/github.com/uncharted-distil/distil-ingest ./vendor/github.com/uncharted-distil/distil-ingest

# make sure we only pick up deps from this project
rm -rf $GOPATH/src/github.com/uncharted-distil/distil-compute/vendor
rm -rf $GOPATH/src/github.com/uncharted-distil/distil-ingest/vendor
