.PHONY: build build-hook test lint run clean

# Build the main TUI binary
build:
	go build -o bin/copilot-icq ./cmd/copilot-icq

# Build the hook companion binary
build-hook:
	go build -o bin/copilot-icq-hook ./cmd/copilot-icq-hook

# Build all binaries
all: build build-hook

# Run the TUI
run: build
	./bin/copilot-icq

# Run tests
test:
	go test ./... -v

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

# Format code
fmt:
	gofmt -w .

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
