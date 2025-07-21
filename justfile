# gmenu justfile

# Build the main binary
build:
	go build -o bin/gmenu -v ./cmd/main.go

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

# Lint code
lint:
	golangci-lint run ./...

# Format code and remove trailing spaces
fmt:
	go fmt ./...
	find . -name "*.go" -exec sed -i '' 's/[[:space:]]*$//' {} \;

# Development example
dev:
	tree -L 5 | go run ./cmd/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run all checks (format, lint, test)
check: fmt lint test

# Install project dependencies and tools
setup: get-deps
	go mod tidy