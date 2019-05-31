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
	@go vet ./...
	@go list ./... | grep -v /vendor/ | xargs -L1 golint

fmt:
	@go fmt ./...

build: lint
	@go build -i ${LDFLAGS}

build_static:
	env CGO_ENABLED=0 env GOOS=linux GOARCH=amd64 go build ${LDFLAGS}

compile: lint
	@go build ./...

deploy:
	@./deploy.sh

watch:
	@./run.sh

test: build
	@go test ./...

install:
	@npm install -g yarn
	@yarn install
	@go get -u golang.org/x/lint/golint
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/unchartedsoftware/witch
	@dep ensure
