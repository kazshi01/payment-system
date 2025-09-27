# Makefile
.PHONY: dev migrate.up migrate.down keycloak.up keycloak.down redis.up redis.down db.remove

dev:
	go run cmd/api/main.go

migrate.up:
	./db/migrate.sh

migrate.down:
	docker compose rm -sfv db

keycloak.up:
	docker compose up -d keycloak

keycloak.down:
	docker compose rm -sfv keycloak

redis.up:
	docker compose up -d redis

redis.down:
	docker compose rm -sfv redis

db.remove:
	docker compose down -v
