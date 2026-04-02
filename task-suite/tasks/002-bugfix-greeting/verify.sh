#!/usr/bin/env bash
set -euo pipefail

workspace="${1:?workspace path is required}"

cd "${workspace}"
go test ./...

echo "task 002 verification passed"
