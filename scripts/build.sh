#!/bin/sh
# Builds baq, injecting the Google OAuth client from .env (if present) so that
# `baq drive login` does not ask for credentials. `.env` is gitignored.
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

if [ -f "$root/.env" ]; then
	set -a
	# shellcheck disable=SC1091
	. "$root/.env"
	set +a
fi

version="dev"
ldflags="-s -w -X github.com/Nicozilabe/baqueano/cmd.AppVersion=$version"
if [ -n "${BAQUEANO_GOOGLE_CLIENT_ID:-}" ]; then
	ldflags="$ldflags -X github.com/Nicozilabe/baqueano/internal/drive.builtinClientID=$BAQUEANO_GOOGLE_CLIENT_ID"
fi
if [ -n "${BAQUEANO_GOOGLE_CLIENT_SECRET:-}" ]; then
	ldflags="$ldflags -X github.com/Nicozilabe/baqueano/internal/drive.builtinClientSecret=$BAQUEANO_GOOGLE_CLIENT_SECRET"
fi

mkdir -p "$root/dist"
cd "$root"
go build -ldflags "$ldflags" -o dist/baq .

if [ -n "${BAQUEANO_GOOGLE_CLIENT_ID:-}" ]; then
	echo "Built $root/dist/baq (with embedded Google client)"
else
	echo "Built $root/dist/baq (no .env / client id; 'baq drive login' will ask for credentials)"
fi
