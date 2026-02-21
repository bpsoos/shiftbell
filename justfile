dev_image := "shiftbell-dev"
export DEV_IMAGE := dev_image

default:
    just --list

build:
    docker compose build

shell: build
    docker run --rm -it -v {{justfile_dir()}}:/src {{dev_image}}

up: build
    docker compose up

down:
    docker compose down

db-shell:
    docker compose exec -ti postgres /bin/bash

db-psql:
    docker compose exec -ti postgres psql -U testuser testdb

dump-dev-db:
    docker compose exec -ti postgres pg_dump -U testuser -F p --exclude-table=schema_migrations --no-schema -f testdb.dump testdb
    docker compose cp postgres:/testdb.dump .

