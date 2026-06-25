-- Phase 1: minimal schema, no auth/projects

CREATE TABLE IF NOT EXISTS products (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS folders (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    parent_id  INTEGER NOT NULL DEFAULT 0,
    name       TEXT NOT NULL,
    path       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(product_id, parent_id, name)
);

CREATE INDEX IF NOT EXISTS idx_folders_product ON folders(product_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(product_id, parent_id);

CREATE TABLE IF NOT EXISTS environments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    base_url   TEXT NOT NULL,
    variables  TEXT NOT NULL DEFAULT '{}',
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(product_id, name)
);

CREATE TABLE IF NOT EXISTS api_definitions (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id        INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    folder_id         INTEGER NOT NULL DEFAULT 0,
    name              TEXT NOT NULL,
    method            TEXT NOT NULL,
    path              TEXT NOT NULL,
    full_url_template TEXT NOT NULL DEFAULT '',
    headers           TEXT NOT NULL DEFAULT '[]',
    body              TEXT NOT NULL DEFAULT '',
    body_type         TEXT NOT NULL DEFAULT 'json',
    description       TEXT NOT NULL DEFAULT '',
    ai_remark         TEXT NOT NULL DEFAULT '',
    source_record     TEXT NOT NULL DEFAULT '',
    created_at        TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_apis_product ON api_definitions(product_id);
CREATE INDEX IF NOT EXISTS idx_apis_folder ON api_definitions(folder_id);

CREATE TABLE IF NOT EXISTS api_assertions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    api_id     INTEGER NOT NULL REFERENCES api_definitions(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    expression TEXT NOT NULL,
    operator   TEXT NOT NULL DEFAULT 'eq',
    expected   TEXT NOT NULL DEFAULT '',
    enabled    INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS scenarios (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id  INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    folder_id   INTEGER NOT NULL DEFAULT 0,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    env_id      INTEGER REFERENCES environments(id) ON DELETE SET NULL,
    created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS scenario_steps (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    scenario_id   INTEGER NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
    step_order    INTEGER NOT NULL,
    name          TEXT NOT NULL,
    api_id        INTEGER,
    method        TEXT NOT NULL,
    path          TEXT NOT NULL,
    headers       TEXT NOT NULL DEFAULT '[]',
    body          TEXT NOT NULL DEFAULT '',
    extract_rules TEXT NOT NULL DEFAULT '[]',
    assertions    TEXT NOT NULL DEFAULT '[]',
    UNIQUE(scenario_id, step_order)
);

CREATE TABLE IF NOT EXISTS runs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    scenario_id INTEGER REFERENCES scenarios(id) ON DELETE SET NULL,
    api_id      INTEGER REFERENCES api_definitions(id) ON DELETE SET NULL,
    env_id      INTEGER NOT NULL REFERENCES environments(id),
    status      TEXT NOT NULL DEFAULT 'running',
    started_at  TEXT NOT NULL DEFAULT (datetime('now')),
    finished_at TEXT,
    summary     TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS run_steps (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id            INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    step_order        INTEGER NOT NULL,
    name              TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'pending',
    request_snapshot  TEXT NOT NULL DEFAULT '{}',
    response_snapshot TEXT NOT NULL DEFAULT '{}',
    assertion_results TEXT NOT NULL DEFAULT '[]',
    duration_ms       INTEGER NOT NULL DEFAULT 0,
    error_message     TEXT NOT NULL DEFAULT ''
);

-- seed products
INSERT OR IGNORE INTO products (id, name) VALUES (1, 'Trex'), (2, 'Edgen'), (3, 'example');
