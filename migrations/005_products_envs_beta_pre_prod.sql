-- 确保 Trex / Edgen / example 三个产品存在
INSERT OR IGNORE INTO products (id, name) VALUES (1, 'Trex');
INSERT OR IGNORE INTO products (id, name) VALUES (2, 'Edgen');
INSERT OR IGNORE INTO products (id, name) VALUES (3, 'example');
UPDATE products SET name = 'Trex' WHERE id = 1;
UPDATE products SET name = 'Edgen' WHERE id = 2;
UPDATE products SET name = 'example' WHERE id = 3;

-- 统一环境为 BETA / PRE / PROD（替换 dev、staging、trex-dev 等）
UPDATE scenarios SET env_id = NULL WHERE env_id IS NOT NULL;
DELETE FROM environments;

INSERT INTO environments (product_id, name, base_url, variables, is_default) VALUES
(1, 'BETA', 'https://beta-api.trex.xyz', '{}', 0),
(1, 'PRE',  'https://pre-api.trex.xyz',  '{}', 0),
(1, 'PROD', 'https://api.trex.xyz',      '{}', 1),
(2, 'BETA', 'https://api.beta.ospprotocol.xyz/', '{}', 0),
(2, 'PRE',  'https://api.pre.ospprotocol.xyz/',  '{}', 0),
(2, 'PROD', 'https://api.edgen.xyz',      '{}', 1),
(3, 'BETA', 'http://localhost:3002',      '{}', 1),
(3, 'PRE',  'http://pre-api.example.com', '{}', 0),
(3, 'PROD', 'http://prod-api.example.com','{}', 0);
