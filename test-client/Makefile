

.PHONY: uber_lint
uber_lint: # Run Uber Go style linter (builds plugin if needed)
	$Q echo "Running Uber Go style linter (with golangci-lint)..."
	$Q if [ ! -f "./custom-gcl" ]; then echo "Building custom golangci-lint with uber-go-lint-style plugin..."; golangci-lint custom || exit 1; fi; echo "Running Uber Go style golangci-lint..." ;./custom-gcl run

.PHONY: uber_clean
uber_clean: # Clean Uber Go style linter artifacts
	$Q rm -f custom-gcl*
	$Q echo "Cleaned Uber Go style linter artifacts"
