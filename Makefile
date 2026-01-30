BINARY_NAME=servctl
VERSION=0.1.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w"

.PHONY: all build clean run help

all: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/servctl

build-local:
	go build ${LDFLAGS} -o bin/${BINARY_NAME}-local ./cmd/servctl

clean:
	rm -rf bin/

run: build-local
	./bin/${BINARY_NAME}-local

help:
	@echo "Usage:"
	@echo "  make build         - Build static binary for Linux AMD64"
	@echo "  make build-local   - Build binary for current OS"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make run           - Build and run locally"
