-include .env

DB_URL_TEST?=postgresql://postgres:@localhost:${DB_PORT_TEST}/${DB_NAME}?serverVersion=14&amp;charset=utf8

export DB_URL_TEST
export GH_PAT

.PHONY: migrate_up
migrate_up: ## Run the database migrations
	migrate -path ./migrations -database postgresql://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable up

.PHONY: migrate_down
migrate_down: ## Rollback the last database migration
	migrate -path ./migrations -database postgresql://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable down

.PHONY: migrate_create
migrate_create: ## Create a new database migration
	migrate create -ext sql -dir migrations -seq create_base_structure

.PHONY: test_unit
test_unit: ## Run the unit tests
	go test -race -coverprofile=profile.cov -coverpkg=./... -tags api  ./...

.PHONY: postgres_test_up
postgres_test_up: ## Setup database for integration tests
	docker compose -f ./docker-compose.dev.yaml up --wait postgres-test

.PHONY: postgres_test_down
postgres_test_down: ## Teardown database for integration tests
	docker compose -f ./docker-compose.dev.yaml down postgres-test

.PHONY: test_integration
test_integration: ## Run the integration tests
	go test -race -count 1 -tags integration -coverpkg=./tests/... ./tests/...

.PHONY: test_integration_local
test_integration_local: ## Run the integration tests with local database
	postgres_test_up test_integration postgres_test_down

.PHONY: test_db
test_db: ## Test database
	docker compose -f ./docker-compose.dev.yaml exec postgres-test bash -c 'pg_prove -U postgres -d platform -r --ext .sql /tmp/tests/db'

.PHONY: help
help: ## Display the commands
	@grep -hE '(^[a-zA-Z0-9_-]+:.*##.*$$)|(^##)' $(MAKEFILE_LIST) \
	| awk 'BEGIN{FS=":.*## "}{ \
		if ($$0 ~ /^##/) { gsub(/^## ?/,""); printf "\033[33m%s\033[0m\n", $$0; } \
		else { printf "\033[32m%-30s\033[0m %s\n", $$1, $$2; } \
	}'

.DEFAULT_GOAL = help
