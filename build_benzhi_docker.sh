#!/usr/bin/env bash
set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
image="${1:-benzhi/community-tool-locker:latest}"
platform="${2:-linux/amd64}"

docker build \
  --platform "$platform" \
  --file "$script_dir/benzhi.Dockerfile" \
  --tag "$image" \
  "$script_dir"

printf 'Built image: %s (%s)\n' "$image" "$platform"
