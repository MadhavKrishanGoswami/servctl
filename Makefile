BINARY_NAME=servctl
VERSION=0.1.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w"

.PHONY: all build clean run test test-short test-coverage test-verbose help

all: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/servctl

build-local:
	go build ${LDFLAGS} -o bin/${BINARY_NAME}-local ./cmd/servctl

clean:
	rm -rf bin/
	rm -rf coverage/

run: build-local
	./bin/${BINARY_NAME}-local

# ============================================
# Testing Commands
# ============================================

# Run unit tests only (safe on any platform - Mac, Windows, Linux)
test-short:
	go test ./... -short -count=1

# Run all tests (some may require Linux)
test:
	go test ./... -count=1

# Run tests with verbose output
test-verbose:
	go test ./... -v -count=1

# Run tests with coverage report
test-coverage:
	@mkdir -p coverage
	go test ./... -coverprofile=coverage/coverage.out -covermode=atomic
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report: coverage/coverage.html"

# Run only storage package tests (always safe)
test-storage:
	go test ./internal/storage/... -v -count=1

# Run only preflight package tests
test-preflight:
	go test ./internal/preflight/... -v -short -count=1

help:
	@echo "Usage:"
	@echo ""
	@echo "  Build Commands:"
	@echo "    make build         - Build static binary for Linux AMD64"
	@echo "    make build-local   - Build binary for current OS"
	@echo "    make clean         - Remove build artifacts"
	@echo "    make run           - Build and run locally"
	@echo ""
	@echo "  Test Commands:"
	@echo "    make test-short    - Run unit tests only (safe on any OS)"
	@echo "    make test          - Run all tests"
	@echo "    make test-verbose  - Run all tests with verbose output"
	@echo "    make test-coverage - Run tests and generate coverage report"
	@echo "    make test-storage  - Run storage package tests only"
	@echo "    make test-preflight- Run preflight package tests only"
	@echo ""
	@echo "  See TESTING.md for more information about test categories."
