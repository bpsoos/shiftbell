dev_image := "shiftbell-dev"
prod_image := "shiftbell"
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
    docker compose down -t 1 --remove-orphans

destroy:
    docker compose down -v -t 1 --remove-orphans

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

go *args: build
    docker run --rm \
        -v {{justfile_dir()}}:/src \
        --entrypoint go \
        --workdir /src \
        {{dev_image}} \
        {{args}}

fmt:
    docker run --rm \
        -v {{justfile_dir()}}:/src \
        --entrypoint /bin/sh \
        --workdir /src \
        {{dev_image}} \
        -c "go fmt ./... && go tool templ fmt ."

test:
    just go tool ginkgo run -r -p --randomize-all --randomize-suites --skip-package=test/acceptance

test-ci:
    just go tool ginkgo run -r -v -p --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --skip-package=test/acceptance

test-acceptance:
    just --justfile test/acceptance/justfile test-acceptance

mockery *args:
    docker run --rm \
        -v {{justfile_dir()}}:/src \
        --workdir /src \
        vektra/mockery:v3.7.2 {{args}}

@regenerate-mocks:
    docker run --rm -v {{justfile_dir()}}:/src \
            --entrypoint /bin/sh \
            --workdir /src \
            alpine \
            -c "find /src -type f \( -name mocks_test.go -o -name mocks.go \) -delete"
    just mockery
