#!/bin/sh
# Builds sshm, injecting the Google OAuth client from .env (if present) so that
# `sshm drive login` does not ask for credentials. `.env` is gitignored.
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

if [ -f "$root/.env" ]; then
	set -a
	# shellcheck disable=SC1091
	. "$root/.env"
	set +a
fi

version="dev"
ldflags="-s -w -X github.com/Gu1llaum-3/sshm/cmd.AppVersion=$version"
if [ -n "${SSHM_GOOGLE_CLIENT_ID:-}" ]; then
	ldflags="$ldflags -X github.com/Gu1llaum-3/sshm/internal/drive.builtinClientID=$SSHM_GOOGLE_CLIENT_ID"
fi
if [ -n "${SSHM_GOOGLE_CLIENT_SECRET:-}" ]; then
	ldflags="$ldflags -X github.com/Gu1llaum-3/sshm/internal/drive.builtinClientSecret=$SSHM_GOOGLE_CLIENT_SECRET"
fi

mkdir -p "$root/dist"
cd "$root"
go build -ldflags "$ldflags" -o dist/sshm .

if [ -n "${SSHM_GOOGLE_CLIENT_ID:-}" ]; then
	echo "Built $root/dist/sshm (with embedded Google client)"
else
	echo "Built $root/dist/sshm (no .env / client id; 'sshm drive login' will ask for credentials)"
fi
