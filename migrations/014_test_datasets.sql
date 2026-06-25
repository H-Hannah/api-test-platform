-- 测试数据集：变量、请求体覆盖、与 TC/API 绑定

CREATE TABLE IF NOT EXISTS test_datasets (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id       INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    version          TEXT NOT NULL DEFAULT '',
    requirement_id   TEXT NOT NULL DEFAULT '',
    dataset_key      TEXT NOT NULL DEFAULT '',
    name             TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    tc_refs          TEXT NOT NULL DEFAULT '[]',
    api_bindings     TEXT NOT NULL DEFAULT '[]',
    variables        TEXT NOT NULL DEFAULT '{}',
    headers_override TEXT NOT NULL DEFAULT '[]',
    body_override    TEXT NOT NULL DEFAULT '',
    obtain_type      TEXT NOT NULL DEFAULT 'env',
    obtain_note      TEXT NOT NULL DEFAULT '',
    owner            TEXT NOT NULL DEFAULT 'qa',
    tags             TEXT NOT NULL DEFAULT '[]',
    source           TEXT NOT NULL DEFAULT 'ai',
    created_at       TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_test_datasets_product ON test_datasets(product_id);
CREATE INDEX IF NOT EXISTS idx_test_datasets_req ON test_datasets(product_id, version, requirement_id);

CREATE TABLE IF NOT EXISTS test_data_specs (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id       INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    version          TEXT NOT NULL,
    requirement_id   TEXT NOT NULL,
    requirement_name TEXT NOT NULL DEFAULT '',
    spec_yaml        TEXT NOT NULL DEFAULT '',
    env_keys         TEXT NOT NULL DEFAULT '[]',
    created_at       TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(product_id, version, requirement_id)
);
