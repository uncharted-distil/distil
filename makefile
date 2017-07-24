VERSION=`git describe --tags`
TIMESTAMP=`date +%FT%T%z`

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.timestamp=${TIMESTAMP}"

.PHONY: all

all:
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  build         - build the source code"
	@echo "  test          - test the source code"
	@echo "  lint          - lint the source code"
	@echo "  fmt           - format the source code"
	@echo "  install       - install dependencies"

lint:
	@go vet $(shell glide novendor)
	@go list ./... | grep -v /vendor/ | xargs -L1 golint

fmt:
	@go fmt $(shell glide novendor)

build: lint
	@go build -i ${LDFLAGS}

compile: lint
	@go build $(shell glide novendor)

watch:
	@./run.sh

test: build
	@go test $(shell glide novendor)

protoc:
	@protoc -I api/pipeline/ api/pipeline/pipeline_service.proto --go_out=plugins=grpc:api/pipeline

install:
	@npm install -g yarn
	@yarn install
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/Masterminds/glide
	@go get -u github.com/unchartedsoftware/witch
	@glide install
