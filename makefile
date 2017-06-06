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
	@go vet $(NOVENDOR)
	@go list ./... | grep -v /vendor/ | xargs -L1 golint

fmt:
	@go fmt $(NOVENDOR)

build: lint
	@go build -i

test: build
	@go test $(NOVENDOR)

install:
	npm install -g yarn
	yarn install
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/Masterminds/glide
	@glide install
