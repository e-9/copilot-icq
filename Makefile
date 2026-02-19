.PHONY: build test lint run clean

# Build the main TUI binary
build:
	go build -o bin/copilot-icq ./cmd/copilot-icq

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
