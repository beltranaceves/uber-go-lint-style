.PHONY: help

# Default target - build plugin if needed and run lint
.DEFAULT_GOAL := lint

# Run linter (builds plugin if needed)
lint:
	@if [ ! -f "./custom-gcl" ]; then \
		echo "Building custom golangci-lint with uber-go-lint-style plugin..."; \
		golangci-lint custom || exit 1; \
	fi
	@./custom-gcl run

# View help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  make             Build plugin (if needed) and run linter"
	@echo "  make clean       Remove cached plugin binary"
	@echo ""
	@echo "Examples:"
	@echo "  make             # First run builds plugin, subsequent runs are fast"
	@echo "  make clean       # Reset and rebuild plugin next time"

.PHONY: clean
clean:
	@rm -f custom-gcl*
	@echo "Cleaned custom linter artifacts"
