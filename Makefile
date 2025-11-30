# ==============================================================================
# Makefile for Project "Link checker"
#
# Description:
#   This Makefile handles the build process for the link checker application,
#   including compilation, testing, packaging, and cleanup.
#
# Usage:
#   make all          - Run all checks (fmt, lint, test, build)
#   make build        - Build the application binary
#   make run          - Run the application
#   make clean        - Clean build artifacts
#   make fmt          - Format code
#   make lint         - Run golangci-lint
#   make test         - Run all tests
#   make test-cover   - Run tests with coverage
#   make test-html    - Generate HTML coverage report
#   make test-verbose - Run tests with verbose output
#   make test-clean   - Clean test artifacts
#
# Author: Bersnakx
# ==============================================================================

SRCDIR := ./cmd/main.go
BINARY_NAME := linkchecker
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html

.PHONY: all build run clean fmt lint test test-cover test-html test-verbose test-clean

# Run all checks: format, lint, test, and build
all: fmt lint test build
	@echo "All checks passed!"

# Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(SRCDIR)
	@echo "Build complete: $(BINARY_NAME)"

# Run the application
run:
	go run $(SRCDIR)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -rf dist/
	@echo "Clean complete"


# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run
	@echo "Linting complete"

# Run all tests
test:
	go test ./...

test-verbose:
	go test -v ./...

test-cover:
	go test -coverprofile=$(COVERAGE_OUT) -coverpkg=./internal/... ./internal/...

test-coverage: test-cover
	@echo "Coverage by package:"
	@go tool cover -func=$(COVERAGE_OUT)
	@echo "Total coverage:"
	@go tool cover -func=$(COVERAGE_OUT) | grep total

test-html: test-cover
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report saved to $(COVERAGE_HTML)"

test-clean:
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
