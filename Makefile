include .envrc

# ====================================================================== #
# HELPERS
# ====================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

# ====================================================================== #
# DEVELOPMENT
# ====================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## run/api-with-smtp: run the cmd/api application with all the arguments
.PHONY: run/api-with-smtp
run/api-with-smtp:
	go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN} --smtp-username ${GREENLIGHT_SMTP_USERNAME} --smtp-password ${GREENLIGHT_SMTP_PASSWORD} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# ====================================================================== #
# QUALITY CONTROL
# ====================================================================== #

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies'
	go mod tidy

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

# ====================================================================== #
# BUILD
# ====================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -o=./bin/api ./cmd/api