-- BDD 功能：从 PRD + 设计方案生成，作为 MR 核对的验收锚点

CREATE TABLE IF NOT EXISTS bdd_features (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id   INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    user_story   TEXT NOT NULL DEFAULT '',
    prd_text     TEXT NOT NULL DEFAULT '',
    design_text  TEXT NOT NULL DEFAULT '',
    gherkin      TEXT NOT NULL DEFAULT '',
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_bdd_features_product ON bdd_features(product_id);
