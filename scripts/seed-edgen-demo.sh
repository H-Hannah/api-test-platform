#!/usr/bin/env bash
# Edgen 全链路演示数据（BDD + 接口场景 + MR-EDGEN-DEMO）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
go run ./cmd/seed-edgen-demo
