.PHONY: help clean profile

# Default target
.DEFAULT_GOAL := profile

# ---- Config ----
PROFILE ?= cpu.out
TOPN    ?= 10
UNIT    ?= ms

# Capture forwarded args (like ./... -run TestFoo)
ARGS := $(filter-out $@,$(MAKECMDGOALS))

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

# Clean profile output
clean:
	@rm -f $(PROFILE)
	@echo "Cleaned profiling artifacts"

# Help
help:
	@echo "Usage: make [target] [go test args]"
	@echo ""
	@echo "Targets:"
	@echo "  make                 Run go test with profiling (default)"
	@echo "  make clean           Remove profiling output"
	@echo ""
	@echo "Examples:"
	@echo "  make ./...                           # Profile all packages"
	@echo "  make ./pkg/foo -run TestX            # Profile specific test"
	@echo "  make ./... -count=1                  # Disable caching"
	@echo "  make ./... -bench=.                  # Profile benchmarks"