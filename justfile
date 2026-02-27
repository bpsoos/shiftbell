dev_image := "shiftbell-dev"
prod_image := "shiftbell"
prod_migrate_image := "shiftbell-migrate"
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
    docker compose exec -ti postgres \
        pg_dump \
        -U testuser \
        -F p \
        --exclude-table=schema_migrations \
        --no-schema \
        -f testdb.dump \
        testdb
    docker compose cp postgres:/testdb.dump .

build-prod:
    docker buildx build --load --platform=linux/amd64 -f Dockerfile.migrate -t {{prod_migrate_image}} .
    docker buildx build --load --platform=linux/amd64 -f Dockerfile.serve -t {{prod_image}} .
    docker save -o out/{{prod_migrate_image}}.tar {{prod_migrate_image}}:latest
    gzip -f out/{{prod_migrate_image}}.tar
    docker save -o out/{{prod_image}}.tar {{prod_image}}:latest
    gzip -f out/{{prod_image}}.tar

fmt:
    docker run --rm -it \
        -v {{justfile_dir()}}:/src \
        --entrypoint /bin/sh \
        --workdir /src \
        {{dev_image}} \
        -c "go fmt ./... && go tool templ fmt ."
