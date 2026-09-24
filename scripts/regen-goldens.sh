#!/usr/bin/env bash
# Regenerates mjml/testdata/*.html (and *.error) with the MJML version pinned
# in mjml/testdata/reference/package.json. Requires Node.js and npm.
set -euo pipefail

reference="$(cd "$(dirname "$0")/.." && pwd)/mjml/testdata/reference"

npm ci --prefix "$reference" --no-audit --no-fund --loglevel=error
node "$reference/render.mjs"
