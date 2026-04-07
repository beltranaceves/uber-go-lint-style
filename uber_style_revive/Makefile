.PHONY: help build test test-verbose lint-example clean

help:
	@echo "Uber Go Style Revive - Build and Test Targets"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build the CLI tool"
	@echo "  test           Run the test suite"
	@echo "  test-verbose   Run tests with verbose output"
	@echo "  lint-example   Lint the example testdata files"
	@echo "  clean          Remove build artifacts"

build:
	@echo "Building uber-go-lint CLI..."
	go build -o uber-go-lint ./cmd/uber-go-lint
	@echo "Build complete: ./uber-go-lint"

test:
	@echo "Running test suite..."
	go test ./rules -v

test-verbose:
	@echo "Running test suite with verbose output..."
	go test ./rules -v -args -verbose

lint-example:
	@echo "Linting positive example (should have violations)..."
	./uber-go-lint testdata/atomic/positive_test.go
	@echo ""
	@echo "Linting negative example (should pass)..."
	./uber-go-lint testdata/atomic/negative_test.go

clean:
	@echo "Cleaning up..."
	rm -f uber-go-lint
	@echo "Clean complete"

.PHONY: help build test test-verbose lint-example clean
