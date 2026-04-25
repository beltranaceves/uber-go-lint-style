.PHONY: help clean profile coverage

# Default target
.DEFAULT_GOAL := profile

# ---- Config ----
PROFILE ?= cpu.out
TOPN    ?= 10
UNIT    ?= ms
COVERAGE_OUT ?= coverage.txt
COVERAGE_FUNC ?= coverage.func.txt
COVERAGE_HTML ?= coverage.html

# Capture forwarded args (like ./... -run TestFoo)
ARGS := $(filter-out help clean profile coverage,$(MAKECMDGOALS))

# Absorb unknown targets so they can act as args
%:
	@:

# Run tests with profiling + show top N
profile:
	@echo "Running go test with CPU profiling..."
	@go test -cpuprofile=$(PROFILE) $(ARGS) >/dev/null || exit 1
	@echo ""
	@echo "Top $(TOPN) CPU hotspots:"
	@go tool pprof -top -nodecount=$(TOPN) -unit=$(UNIT) $(PROFILE)

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -covermode=atomic -coverpkg=./... -coverprofile=$(COVERAGE_OUT) $(ARGS) ./... || true
	@if [ -f $(COVERAGE_OUT) ]; then \
		echo ""; \
		echo "Function-level coverage:"; \
		go tool cover -func=$(COVERAGE_OUT) | tee $(COVERAGE_FUNC); \
		echo ""; \
		go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML); \
		echo "HTML coverage report: $(COVERAGE_HTML)"; \
	else \
		echo "No coverage data generated"; \
	fi

# Clean profile output
clean:
	@rm -f $(PROFILE) $(COVERAGE_OUT) $(COVERAGE_FUNC) $(COVERAGE_HTML)
	@echo "Cleaned profiling and coverage artifacts"

# Help
help:
	@echo "Usage: make [target] [go test args]"
	@echo ""
	@echo "Targets:"
	@echo "  make                 Run go test with profiling (default)"
	@echo "  make coverage        Run go test with coverage reporting"
	@echo "  make clean           Remove profiling and coverage output"
	@echo ""
	@echo "Examples:"
	@echo "  make ./...                           # Profile all packages"
	@echo "  make ./pkg/foo -run TestX            # Profile specific test"
	@echo "  make ./... -count=1                  # Disable caching"
	@echo "  make ./... -bench=.                  # Profile benchmarks"
	@echo "  make coverage ./...                  # Generate coverage report"

# uber-go-lint-style plugin targets


.PHONY: uber_lint
uber_lint: # Run Uber Go style linter (builds plugin if needed)
	$Q echo "Running Uber Go style linter (with golangci-lint)..."
	$Q if [ ! -f "./custom-gcl" ]; then echo "Building custom golangci-lint with uber-go-lint-style plugin..."; golangci-lint custom || exit 1; fi; echo "Running Uber Go style golangci-lint..." ;./custom-gcl run --config .golangci.uber_style.yml

.PHONY: uber_clean
uber_clean: # Clean Uber Go style linter artifacts
	$Q rm -f custom-gcl*
	$Q echo "Cleaned Uber Go style linter artifacts"
