include .env
export

GOOSE = go tool goose -dir migrations postgres "$(DATABASE_URL)"

.PHONY: db-up db-down migrate migrate-down migrate-status migrate-create

db-up:
	docker compose up -d --wait

db-down:
	docker compose down

migrate:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

migrate-create:
	go tool goose -dir migrations -s create $(name) sql
