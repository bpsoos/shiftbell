#!/usr/bin/env bash

set -euo pipefail

version="${1:?HTMX version is required}"
js_sha256="${2:?HTMX JavaScript SHA-256 is required}"

wget -q "https://cdn.jsdelivr.net/npm/htmx.org@$version/dist/htmx.min.js" -O internal/static/assets/htmx.min.js
wget -q "https://cdn.jsdelivr.net/npm/htmx.org@$version/LICENSE" -O notices/htmx-0BSD.txt
printf '%s  %s\n' \
    "$js_sha256" internal/static/assets/htmx.min.js |
    sha256sum -c -
