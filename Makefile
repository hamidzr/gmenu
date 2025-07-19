build:
	go build -o bin/gmenu -v ./cmd/main.go

build-all:
	bash ./scripts/build.sh

get-deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

test:
	gotestsum -- ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

dev:
	tree -L 5 | go run ./cmd/main.go
