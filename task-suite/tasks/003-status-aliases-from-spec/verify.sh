#!/usr/bin/env bash
set -euo pipefail

workspace="${1:?workspace path is required}"

test -f "${workspace}/internal/status/status_test.go"

cd "${workspace}"
go test ./...

output="$(go run ./cmd/check)"
expected=$'active\ninactive\nactive\nunknown'

if [[ "${output}" != "${expected}" ]]; then
	echo "unexpected cmd/check output:" >&2
	printf '%s\n' "${output}" >&2
	exit 1
fi

echo "task 003 verification passed"
