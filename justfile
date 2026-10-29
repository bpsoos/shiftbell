dev_image := "shiftbell-dev"
prod_image := "shiftbell"
demo_image := "shiftbell-demo"
golangci_lint_image := "golangci/golangci-lint:v2.15.0"
bootstrap_version := "5.3.8"
bootstrap_css_sha256 := "05b86c3a7fef0df5874fd856c683ec09cb17dceeab36da061ebd1667796fea79"
bootstrap_js_sha256 := "e4fd49181388c48ec5040bd3fe66f57c29c8e67fcd8502b3354b96ec7ab47cc7"
htmx_version := "2.0.11"
htmx_js_sha256 := "d6fdc75f204e6bdefa99b69bf1e6d4ac69b8a364f77929f45c13476b4000f717"
export DEV_IMAGE := dev_image
export LOCAL_UID := `id -u`
export LOCAL_GID := `id -g`

default:
    just --list

build:
    docker compose build -q

build-prod:
    docker build -f Dockerfile.serve -t {{prod_image}} .

build-demo: build-prod
    docker build -f Dockerfile.demo --build-arg PROD_IMAGE={{prod_image}} -t {{demo_image}} .

shell: build (_prepare-cache "shiftbell-go-cache")
    docker run --rm -it \
        --user "$LOCAL_UID:$LOCAL_GID" \
        -e HOME=/tmp \
        -e GOCACHE=/cache/go-build \
        -e GOMODCACHE=/cache/go-mod \
        -v shiftbell-go-cache:/cache \
        -v {{justfile_dir()}}:/src {{dev_image}}

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

go *args: build (_prepare-cache "shiftbell-go-cache")
    docker run --rm \
        --user "$LOCAL_UID:$LOCAL_GID" \
        -e HOME=/tmp \
        -e GOCACHE=/cache/go-build \
        -e GOMODCACHE=/cache/go-mod \
        -v shiftbell-go-cache:/cache \
        -v {{justfile_dir()}}:/src \
        --entrypoint go \
        --workdir /src \
        {{dev_image}} \
        {{args}}

@fmt:
    just go tool templ fmt
    just golangci-lint fmt

test:
    just go tool ginkgo run -r -p --randomize-all --randomize-suites --skip-package=test/acceptance

test-ci:
    just go tool ginkgo run -r -v -p --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --skip-package=test/acceptance

test-acceptance:
    just --justfile test/acceptance/justfile test

@lint:
    just golangci-lint run

@generate-templ:
    just go tool templ generate

fetch-bootstrap: build
    docker run --rm \
        --user "$(id -u):$(id -g)" \
        -v {{justfile_dir()}}:/src \
        --workdir /src \
        --entrypoint /bin/bash \
        {{dev_image}} \
        scripts/fetch-bootstrap-assets.sh \
        "{{bootstrap_version}}" \
        "{{bootstrap_css_sha256}}" \
        "{{bootstrap_js_sha256}}"

fetch-htmx: build
    docker run --rm \
        --user "$(id -u):$(id -g)" \
        -v {{justfile_dir()}}:/src \
        --workdir /src \
        --entrypoint /bin/bash \
        {{dev_image}} \
        scripts/fetch-htmx-assets.sh \
        "{{htmx_version}}" \
        "{{htmx_js_sha256}}"

@golangci-lint *args: (_prepare-cache "shiftbell-golangci-cache")
    docker run --rm \
        --user "$LOCAL_UID:$LOCAL_GID" \
        -e HOME=/tmp \
        -v "{{justfile_directory()}}:/app" \
        -v "shiftbell-golangci-cache:/cache" \
        -w /app \
        -e GOLANGCI_LINT_CACHE=/cache/golangci-lint \
        -e GOCACHE=/cache/go-build \
        -e GOMODCACHE=/cache/go-mod \
        {{golangci_lint_image}} \
        golangci-lint {{args}}

mockery *args: (_prepare-cache "shiftbell-go-cache")
    docker run --rm \
        --user "$LOCAL_UID:$LOCAL_GID" \
        -e HOME=/tmp \
        -e GOCACHE=/cache/go-build \
        -e GOMODCACHE=/cache/go-mod \
        -v shiftbell-go-cache:/cache \
        -v {{justfile_dir()}}:/src \
        --workdir /src \
        vektra/mockery:v3.8.1 {{args}}

@regenerate-mocks:
    docker run --rm -v {{justfile_dir()}}:/src \
            --user "$LOCAL_UID:$LOCAL_GID" \
            --entrypoint /bin/sh \
            --workdir /src \
            alpine \
            -c "find /src -type f \( -name mocks_test.go -o -name mocks.go \) -delete"
    just mockery

[private]
_prepare-cache volume:
    docker run --rm \
        -v "{{volume}}:/cache" \
        alpine chown -R "$LOCAL_UID:$LOCAL_GID" /cache
