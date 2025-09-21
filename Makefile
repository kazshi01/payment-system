# Makefile
.PHONY: migrate.up migrate.remove keycloak.up keycloak.remove db.down

migrate.up:
	./db/migrate.sh

migrate.remove:
	docker compose rm -sfv db

keycloak.up:
	docker compose up -d keycloak

keycloak.remove:
	docker compose rm -sfv keycloak

db.down:
	docker compose down -v
