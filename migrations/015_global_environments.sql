-- 环境全局化：去掉 product_id，同名环境合并（保留 variables 最丰富的一条）

CREATE TABLE _env_map (old_id INTEGER PRIMARY KEY, new_id INTEGER NOT NULL);

CREATE TABLE environments_new (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL UNIQUE,
    base_url   TEXT NOT NULL,
    variables  TEXT NOT NULL DEFAULT '{}',
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO environments_new (name, base_url, variables, is_default, created_at)
SELECT e.name, e.base_url, e.variables, e.is_default, e.created_at
FROM environments e
INNER JOIN (
    SELECT name, id FROM environments e1
    WHERE id = (
        SELECT id FROM environments e2
        WHERE e2.name = e1.name
        ORDER BY LENGTH(e2.variables) DESC, e2.product_id DESC, e2.id ASC
        LIMIT 1
    )
) pick ON pick.id = e.id;

INSERT INTO _env_map (old_id, new_id)
SELECT e.id, n.id
FROM environments e
JOIN environments_new n ON n.name = e.name;

UPDATE scenarios
SET env_id = (SELECT new_id FROM _env_map WHERE old_id = scenarios.env_id)
WHERE env_id IS NOT NULL;

UPDATE runs
SET env_id = (SELECT new_id FROM _env_map WHERE old_id = runs.env_id)
WHERE env_id IS NOT NULL;

DROP TABLE environments;
ALTER TABLE environments_new RENAME TO environments;

UPDATE environments SET is_default = 0;
UPDATE environments SET is_default = 1 WHERE name = 'PROD';

DROP TABLE _env_map;
