#!/bin/bash

# Build custom golangci-lint with uber-go-lint-style plugin if needed
if [ ! -f "./custom-gcl" ]; then
    echo "Building custom golangci-lint binary with uber-go-lint-style plugin..."
    golangci-lint custom || exit 1
    echo "Custom binary built: ./custom-gcl"
fi

# Run the linter with the custom binary
./custom-gcl "$@"
