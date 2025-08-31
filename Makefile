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
	@cd frontend && npm run dev

## db:	Start the database using Docker Compose
.PHONY: db
db:
	@docker compose up --build --remove-orphans

## migrate-up:	Run database migrations up
.PHONY: migrate-up
migrate-up:
	@cd migrations && go run ./up/up.go --dsn $(POSTGRES_URL) --dir .

## migrate-down:	Run database migrations down
.PHONY: migrate-down
migrate-down:
	@cd migrations && go run ./down/down.go --dsn $(POSTGRES_URL) --dir .

.PHONY: tests
tests: tests-api tests-frontend

.PHONY: tests-api
tests-api:
	@./test-api.sh

.PHONY: tests-frontend
tests-frontend:
	@set -e; \
	cd frontend; \
	( npm run dev >/dev/null 2>&1 & echo $$! > ../.front.pid ); \
	cd ..; \
	echo "Attente du front sur :3000..."; \
	for i in $$(seq 1 60); do \
		if curl -sf http://localhost:3000/ >/dev/null; then \
			echo "Frontend prêt"; break; \
		fi; \
		sleep 0.5; \
	done; \
	cd frontend && npx cypress run; \
	cd ..; \
	if [ -f .front.pid ]; then \
		kill $$(cat .front.pid) 2>/dev/null || true; \
		rm -f .front.pid; \
	fi
