#!/usr/bin/env bash
# 清除平台测试数据：接口、用例、场景、报告、BDD/测试数据规格等；保留产品与运行环境。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DB="${DB_PATH:-$ROOT/data/platform.db}"

if [[ ! -f "$DB" ]]; then
  echo "数据库不存在: $DB" >&2
  exit 1
fi

sqlite3 "$DB" <<'SQL'
PRAGMA foreign_keys = ON;
DELETE FROM run_steps;
DELETE FROM runs;
DELETE FROM scenario_steps;
DELETE FROM scenarios;
DELETE FROM test_datasets;
DELETE FROM test_data_specs;
DELETE FROM bdd_features;
DELETE FROM api_assertions;
DELETE FROM api_definitions;
DELETE FROM folders;
DELETE FROM sqlite_sequence WHERE name IN (
  'folders', 'api_definitions', 'api_assertions',
  'scenarios', 'scenario_steps', 'runs', 'run_steps',
  'test_datasets', 'test_data_specs', 'bdd_features'
);
SQL

echo "✅ 已清除测试数据（保留 products / environments）"
sqlite3 "$DB" <<'SQL'
.mode column
.headers on
SELECT 'products' AS tbl, COUNT(*) AS cnt FROM products
UNION ALL SELECT 'environments', COUNT(*) FROM environments
UNION ALL SELECT 'folders', COUNT(*) FROM folders
UNION ALL SELECT 'api_definitions', COUNT(*) FROM api_definitions
UNION ALL SELECT 'test_datasets', COUNT(*) FROM test_datasets
UNION ALL SELECT 'test_data_specs', COUNT(*) FROM test_data_specs
UNION ALL SELECT 'scenarios', COUNT(*) FROM scenarios
UNION ALL SELECT 'runs', COUNT(*) FROM runs
UNION ALL SELECT 'bdd_features', COUNT(*) FROM bdd_features;
SQL
