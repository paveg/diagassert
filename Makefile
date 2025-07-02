.PHONY: test lint fmt clean coverage help

# Default target
all: test

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -f coverage.txt coverage.html

# Install development dependencies
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  test          - Run tests"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  help          - Show this help message"