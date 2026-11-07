#!/bin/sh
set -eu

exec biome format \
    --stdin-file-path="$TEMPL_PRETTIER_FILENAME" \
    --html-formatter-enabled=true \
    --html-formatter-indent-script-and-style=true
