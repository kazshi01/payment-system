# Makefile
.PHONY: migrate.up migrate.down

migrate.up:
	./db/migrate.sh

migrate.down:
	docker compose down -v
