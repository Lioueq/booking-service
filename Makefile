include .env

.PHONY: up down tests seed

up:
	docker compose up -d --build

down:
	docker compose down

tests:
	go test ./...

seed:
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < db/seed.sql

