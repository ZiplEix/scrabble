ifneq (,$(wildcard ./api/.env))
	include ./api/.env
	export
endif

.ONESHELL:

.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## air:	Run the application with live reload
.PHONY: air
air:
	@cd api && air

## front:	Run the frontend application in development mode
.PHONY: front
front:
	@cd frontend && bun run dev

## admin:	Run the admin interface in development mode
.PHONY: admin
admin:
	@cd admin-dashboard && bun run dev -- --host

## db:	Start the database using Docker Compose
.PHONY: db
db:
	@docker compose up --build --remove-orphans

## migrate-up:	Run database migrations up
.PHONY: migrate-up
migrate-up:
	@cd api/migrations && go run ./up/up.go --dsn $(POSTGRES_URL) --dir .

## migrate-down:	Run database migrations down
.PHONY: migrate-down
migrate-down:
	@cd api/migrations && go run ./down/down.go --dsn $(POSTGRES_URL) --dir .

## recalculate-ips:	Recalculate all user IPS from last 10 games
.PHONY: recalculate-ips
recalculate-ips:
	@cd api && go run cmd/recalculate-ips/main.go

## recalculate-achievements:	Recalculate and unlock all achievements retroactively for all players
.PHONY: recalculate-achievements
recalculate-achievements:
	@cd api && go run cmd/recalculate-achievements/main.go

## tests:	Run all tests (API and frontend)
.PHONY: tests
tests: tests-api tests-frontend

## tests-api:	Run API tests
.PHONY: tests-api
tests-api:
	@./test-api.sh

## tests-frontend:	Run frontend tests
.PHONY: tests-frontend
tests-frontend:
	@./test-front.sh
