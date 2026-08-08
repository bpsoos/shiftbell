#!/usr/bin/env bash

set -euo pipefail

version="${1:?Bootstrap version is required}"
css_sha256="${2:?Bootstrap CSS SHA-256 is required}"
js_sha256="${3:?Bootstrap JavaScript SHA-256 is required}"

wget -q "https://cdn.jsdelivr.net/npm/bootswatch@$version/dist/sandstone/bootstrap.min.css" -O internal/static/assets/bootstrap.min.css
wget -q "https://cdn.jsdelivr.net/npm/bootstrap@$version/dist/js/bootstrap.bundle.min.js" -O internal/static/assets/bootstrap.bundle.min.js
wget -q "https://cdn.jsdelivr.net/npm/bootstrap@$version/LICENSE" -O notices/bootstrap-MIT.txt
printf '%s  %s\n' \
    "$css_sha256" internal/static/assets/bootstrap.min.css \
    "$js_sha256" internal/static/assets/bootstrap.bundle.min.js |
    sha256sum -c -
