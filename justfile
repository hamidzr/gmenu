# gmenu justfile

# Build the main binary
[parallel]
build: build-go build-zig

build-go:
	go build -o bin/gmenu -v ./cmd

build-zig:
	just -f ./zig/justfile build

# Build for multiple platforms (Darwin amd64/arm64)
build-all:
	bash ./scripts/build.sh

# Install dependencies
get-deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install gotest.tools/gotestsum@latest

# Run tests
test:
	gotestsum --format=short -- -timeout=30s ./...

test-race:
	gotestsum --format=short -- -race -timeout=30s ./...

# Run visual GUI tests (shows actual GUI windows)
test-visual:
	./run_visual_tests.sh

# Lint code
lint:
	golangci-lint run ./...

# Format code and remove trailing spaces
fmt:
	go fmt ./...
	find . -name "*.go" -exec sed -i '' 's/[[:space:]]*$//' {} \;

# Development example
dev:
	tree -L 5 | go run ./cmd

# Run with CPU profiling enabled (writes cpu.pprof)
profile:
	tree -L 5 | go run -tags=pprof ./cmd

# Clean build artifacts
clean:
	rm -rf bin/

# Run all checks (format, lint, test)
check: fmt lint test

# Install project dependencies and tools
setup: get-deps
	go mod tidy

zig CMD:
	just -f ./zig/justfile $(CMD)
