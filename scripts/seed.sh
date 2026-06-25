#!/usr/bin/env bash
# 初始化数据库：表结构 + 环境 + 示例目录树
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export DB_PATH="${DB_PATH:-./data/platform.db}"

echo "📦 Seed 数据库: $DB_PATH"
go run ./cmd/seed
