# Load environment variables
include .envrc

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: Start the API server
.PHONY: run/api
run/api:
	go run ./cmd/api \
		-db-dsn='$(MEDICAL_DB_DSN)' \
		-cors-trusted-origins='$(CORS_TRUSTED_ORIGINS)'

## run/api-no-limit: Start the API server without rate limiting
.PHONY: run/api-no-limit
run/api-no-limit:
	go run ./cmd/api \
		-db-dsn='$(MEDICAL_DB_DSN)' \
		-limiter-enabled=false \
		-cors-trusted-origins='$(CORS_TRUSTED_ORIGINS)'

# ==================================================================================== #
# DEMO
# ==================================================================================== #

## demo/shutdown-request: Send slow request for graceful shutdown demo
.PHONY: demo/shutdown-request
demo/shutdown-request:
	@echo "Sending slow request (10 second delay)..."
	@curl -i http://localhost:4000/v1/slow
	@echo ""
	@echo "Request completed successfully!"

## run/cors-basic: Run the basic CORS demo
.PHONY: run/cors-basic
run/cors-basic:
	go run ./cmd/examples/cors/basic

## run/cors-preflight: Run the preflight CORS demo
.PHONY: run/cors-preflight
run/cors-preflight:
	go run ./cmd/examples/cors/preflight

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

## db/psql: Connect to the database using psql
.PHONY: db/psql
db/psql:
	psql '$(MEDICAL_DB_DSN)'

# ==================================================================================== #
# MIGRATIONS
# ==================================================================================== #

## db/migrations/new name=$1 : Create new migration files
.PHONY: db/migrations/new
db/migrations/new:
	@test -n '$(name)' || (echo 'Usage: make db/migrations/new name=create_table' && exit 1)
	@echo 'Creating migration files for $(name)...'
	migrate create -seq -ext=.sql -dir=./migrations $(name)

## db/migrations/up: Apply all up migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database '$(MEDICAL_DB_DSN)' up

## db/migrations/down: Apply all down migrations (revert all)
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Reverting all migrations...'
	migrate -path ./migrations -database '$(MEDICAL_DB_DSN)' down

## db/migrations/goto version=$1: Go to specified migration version
.PHONY: db/migrations/goto
db/migrations/goto:
	@echo 'Going to migration version ${version}...'
	migrate -path ./migrations -database ${MEDICAL_DB_DSN} goto ${version}

## db/migrations/fix version=$1: Force the migration to a specific version
.PHONY: db/migrations/fix
db/migrations/fix:
	@test -n '$(version)' || (echo 'Usage: make db/migrations/fix version=1' && exit 1)
	@echo 'Forcing migration version to $(version)...'
	migrate -path ./migrations -database '$(MEDICAL_DB_DSN)' force $(version)