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

build_static:
	env CGO_ENABLED=0 env GOOS=linux GOARCH=amd64 go build ${LDFLAGS}

compile: lint
	@go build $(shell glide novendor)

deploy:
	@./deploy.sh

watch:
	@./run.sh

test: build
	@go test $(shell glide novendor)

proto:
	@protoc -I /usr/local/include -I api/pipeline api/pipeline/*.proto --go_out=plugins=grpc:api/pipeline

install:
	@npm install -g yarn
	@yarn install
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/unchartedsoftware/witch
	@dep ensure
