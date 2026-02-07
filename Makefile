BINARY_NAME=servctl
VERSION=0.1.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w"

.PHONY: all build clean run test test-short test-coverage docker-test docker-shell help

all: build

# ============================================
# Build Commands
# ============================================

# Build static binary for Linux AMD64 (for Docker testing)
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/servctl
	@echo "Built: bin/${BINARY_NAME} (Linux AMD64)"

# Build binary for current OS (Mac)
build-local:
	go build ${LDFLAGS} -o bin/${BINARY_NAME}-local ./cmd/servctl
	@echo "Built: bin/${BINARY_NAME}-local ($(shell go env GOOS)/$(shell go env GOARCH))"

clean:
	rm -rf bin/
	rm -rf coverage/

run: build-local
	./bin/${BINARY_NAME}-local

# ============================================
# Unit Tests (Run on Mac)
# ============================================

# Run unit tests only (safe on any platform)
test-short:
	go test ./... -short -count=1

# Run all tests
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

# ============================================
# Docker Testing (Ubuntu 22.04 Simulation)
# ============================================

# Build test container
docker-build: build
	docker build -f build/Dockerfile -t servctl-test .
	@echo ""
	@echo "✅ Docker test image built: servctl-test"

# Run full test suite in container
docker-test: docker-build
	@echo ""
	docker run --rm --privileged servctl-test
	@echo ""
	@echo "✅ Docker tests complete!"

# Interactive shell in test container
docker-shell: docker-build
	@echo "Opening Ubuntu 22.04 shell with servctl..."
	@echo "Commands to try:"
	@echo "  ./servctl -version"
	@echo "  sudo ./servctl -preflight"
	@echo "  sudo ./simulate-disks.sh  (create virtual disks)"
	@echo ""
	docker run --rm -it --privileged servctl-test /bin/bash

# Quick version check only
docker-quick: build
	docker run --rm -v $(PWD)/bin:/app ubuntu:22.04 /app/servctl -version

# Full Docker-in-Docker test (can run docker-compose!)
docker-full: build
	docker build -f build/Dockerfile.dind -t servctl-dind .
	@echo ""
	@echo "Starting Docker-in-Docker container..."
	docker run --rm --privileged -it servctl-dind

help:
	@echo "Usage:"
	@echo ""
	@echo "  Build:"
	@echo "    make build         - Build Linux binary"
	@echo "    make build-local   - Build Mac binary"
	@echo "    make clean         - Remove build artifacts"
	@echo ""
	@echo "  Test (Mac):"
	@echo "    make test-short    - Run unit tests (fast)"
	@echo "    make test-coverage - Generate coverage report"
	@echo ""
	@echo "  Test (Docker):"
	@echo "    make docker-test   - Run tests in Ubuntu container"
	@echo "    make docker-shell  - Interactive Ubuntu shell"
	@echo "    make docker-full   - Full test with Docker-in-Docker"
	@echo ""
	@echo "  Quick Start:"
	@echo "    make test-short && make docker-test"
	@echo ""
