.PHONY: tidy
tidy:
	@go mod tidy && go mod vendor

.PHONY: run
run: ## Run Interaction API application
	@go run . serve -c file://.env