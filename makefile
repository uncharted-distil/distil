version=0.1.0

.PHONY: all

NOVENDOR := $(shell glide novendor)

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
	@go build -i

run: lint
	@go build $(shell glide novendor)
	@go run main.go

test: build
	@go test $(shell glide novendor)

install:
	@npm install -g yarn
	@yarn install
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/Masterminds/glide
	@glide install
