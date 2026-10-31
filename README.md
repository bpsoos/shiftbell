# shiftbell

A self-hosted chore manager.

[Try the live demo](https://demo.shiftbell.dev/).

## features

- Create and complete one-off chores.
- Define recurring chores, where completion automatically creates the next one.
- Keep a completed chore history.

## audience

Intended for self-hosting in a homelab. Designed around my personal use and preferences.

## stack

Go+htmx and sqlite, packaged as a single binary.

## quick start

The examples below assume Linux.

Create a data folder in your chosen installation directory:

```sh
mkdir -p data
```

The data folder holds the sqlite database. Use a persistent local folder.

Set `APP_TIMEZONE` to your timezone. Open [localhost:8080](http://localhost:8080) once started.

### docker

```sh
docker run -d --name shiftbell --restart unless-stopped \
  --user "$(id -u):$(id -g)" \
  -p 8080:80 \
  -v "$(pwd)/data:/data" \
  -e DATABASE_FILEPATH=/data/app.db \
  -e APP_TIMEZONE=Europe/Budapest \
  bpsoos/shiftbell:latest
```

### docker compose

Save as `compose.yaml` in that directory:

```yaml
services:
  shiftbell:
    image: bpsoos/shiftbell:latest
    restart: unless-stopped
    user: "${LOCAL_UID}:${LOCAL_GID}"
    ports:
      - "8080:80"
    environment:
      DATABASE_FILEPATH: /data/app.db
      APP_TIMEZONE: Europe/Budapest
    volumes:
      - "./data:/data"
```

Run from that folder. The user and group IDs let the container write to the data directory created earlier:

```sh
LOCAL_UID=$(id -u) LOCAL_GID=$(id -g) docker compose up -d
```

### binary from source

Requires git and go 1.27 or newer.

Clone the repository and build:

```sh
git clone https://github.com/bpsoos/shiftbell.git
cd shiftbell
go tool templ generate
CGO_ENABLED=0 go build -o shiftbell ./cmd/serve
```

Start the binary with `APP_PORT=8080` (the default is 80):

```sh
mkdir -p data
DATABASE_FILEPATH=./data/app.db APP_TIMEZONE=Europe/Budapest APP_PORT=8080 ./shiftbell
```

## security

> [!IMPORTANT]
> Shiftbell does not provide built-in authentication or TLS. It is intended for use on trusted local networks and should not be exposed directly to the public internet.

For remote access, use a reverse proxy with HTTPS and authentication, or a private VPN.

## development

Use the `justfile` (see [just](https://just.systems/man/en/installation.html)) recipes.

Start a local dockerized instance listening on `localhost:8080`:

```sh
just up
```
