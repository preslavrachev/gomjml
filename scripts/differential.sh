#!/usr/bin/env bash
# Compares gomjml with the MJML pinned in mjml/testdata/reference over the
# differential corpus (mjml/differential_test.go), and writes a clustered report
# to mjml/testdata/differential/.cache/report/. Requires Node.js, npm and git.
#
#   scripts/differential.sh           compare with MJML; fail on any change
#   scripts/differential.sh --update  also rewrite the records and mrml goldens
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
reference="$root/mjml/testdata/reference"
cache="$root/mjml/testdata/differential/.cache"

# mjmlio/email-templates states no licence, so it is fetched here rather than committed.
templates_repo=https://github.com/mjmlio/email-templates
templates_commit=a7230b13086c1746524839768fe0bc15a897d960

args=(-diff.results="$cache/results.jsonl")
for arg in "$@"; do
  case "$arg" in
    --update) args+=(-diff.update) ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

mkdir -p "$cache"
templates="$cache/email-templates"
if [ "$(git -C "$templates" rev-parse HEAD 2>/dev/null)" != "$templates_commit" ]; then
  rm -rf "$templates"
  git init --quiet "$templates"
  git -C "$templates" fetch --quiet --depth 1 "$templates_repo" "$templates_commit"
  git -C "$templates" -c advice.detachedHead=false checkout --quiet FETCH_HEAD
fi

npm ci --prefix "$reference" --no-audit --no-fund --loglevel=error

cd "$root/mjml"
go test -run '^TestDifferential$' -count=1 . -args -diff.manifest="$cache/manifest.jsonl"
node "$reference/batch.mjs" "$cache/manifest.jsonl" "$cache/results.jsonl"
go test -run '^TestDifferential$' -count=1 -timeout 30m . -args "${args[@]}"
