# ==========================================
# Bar 108 — Makefile
# ==========================================

# Load .env file so make commands can use env vars
include .env
export

APP_NAME=bar108
CMD_PATH=./cmd/main.go
BIN_PATH=./bin/$(APP_NAME)
# ——— Development ———————————————————————————

## run: start the server (with live .env loaded)
run:
	go run $(CMD_PATH)

## build: compile the binary into ./bin/
build:
	go build -o $(BIN_PATH) $(CMD_PATH)

## clean: remove compiled binaries
clean:
	rm -rf ./bin
## test: run all tests with race detector
test:
	go test -v -race ./...
## test-coverage: run tests and show coverage percentage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# ——— Code quality ——————————————————————————

## fmt: format all Go files
fmt:
	gofmt -w .

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## vet: run go vet (built-in checker)
vet:
	go vet ./...

## check: fmt + vet + lint (run before committing)
check: fmt vet lint

# ——— Database ——————————————————————————————

## db-up: start PostgreSQL via Docker
db-up:
	docker-compose up -d

## db-down: stop PostgreSQL
db-down:
	docker-compose down

## db-logs: show database logs
db-logs:
	docker-compose logs -f db
## migrate: run SQL migrations
migrate:
	docker exec -i bar108_db psql -U $(DB_USER) -d $(DB_NAME) < db/migrations/001_init.sql
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up
migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down
new-migrate:
	migrate create -ext sql -dir db/migrations -seq add_orders_index
## db-shell: open interactive postgres shell
db-shell:
	docker exec -it bar108_db psql -U $(DB_USER) -d $(DB_NAME)
	## sqlc: generate Go code from SQL queries
sqlc:
	sqlc generate

## generate: run all go:generate directives
generate:
	go generate ./...

## mocks: generate all mocks using mockgen
mocks:
	mockgen -source=internal/services/menu_service.go \
		-destination=internal/services/mocks/menu_service_mock.go \
		-package=mocks
	mockgen -source=internal/services/user_services.go \
		-destination=internal/services/mocks/user_service_mock.go \
		-package=mocks
	mockgen -source=internal/repository/order_repository.go \
		-destination=internal/repository/mocks/order_repository_mock.go \
		-package=mocks
	mockgen -source=internal/repository/menu_repository.go \
		-destination=internal/repository/mocks/menu_repository_mock.go \
		-package=mocks
# ——— Help ——————————————————————————————————

## help: print all available commands
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

.PHONY: run build clean fmt lint vet check db-up db-down db-logs help migrate db-shell sqlc migrate-up migrate-down test test-coverage generate mocks