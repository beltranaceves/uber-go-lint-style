.PHONY: help lint lint-revive lint-check install-linters docs

help:
	@echo "Uber Go Lint Style - Development Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  lint              - Run all linters via golangci-lint"
	@echo "  lint-revive       - Run only revive linter"
	@echo "  lint-check        - Check if golangci-lint is installed"
	@echo "  install-linters   - Install golangci-lint"
	@echo "  docs              - Generate style guide documentation"
	@echo "  help              - Show this help message"

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

lint-revive:
	@echo "Running revive linter..."
	@golangci-lint run --linters=revive ./...

lint-check:
	@which golangci-lint > /dev/null || echo "golangci-lint not found. Run 'make install-linters'"

install-linters:
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "golangci-lint installed successfully"

docs:
	@echo "Documentation is generated from style_guide/rules/"
	@echo "See LINTING.md and .github/copilot-instructions.md for development guides"
