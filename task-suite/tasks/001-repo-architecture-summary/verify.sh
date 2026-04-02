#!/usr/bin/env bash
set -euo pipefail

workspace="${1:?workspace path is required}"
summary="${workspace}/architecture-summary.md"

test -f "${summary}"

grep -q "^## Entry Point" "${summary}"
grep -q "^## Config" "${summary}"
grep -q "^## Runner" "${summary}"
grep -q "^## Tooling" "${summary}"
grep -q "^## Request Flow" "${summary}"

grep -q "cmd/mini-agent/main.go" "${summary}"
grep -q "internal/config/config.go" "${summary}"
grep -q "internal/runner/runner.go" "${summary}"
grep -q "internal/tool/echo.go" "${summary}"

echo "task 001 verification passed"
