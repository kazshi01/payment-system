# Makefile
.PHONY: migrate.up migrate.down keycloak.up keycloak.down

migrate.up:
	./db/migrate.sh

migrate.down:
	docker compose rm -sfv db

keycloak.up:
	docker compose up -d keycloak

keycloak.down:
	docker compose rm -sfv keycloak
