-- 三个归属项目：Trex / Edgen / example

INSERT OR IGNORE INTO products (id, name) VALUES (1, 'Trex');
INSERT OR IGNORE INTO products (id, name) VALUES (2, 'Edgen');
INSERT OR IGNORE INTO products (id, name) VALUES (3, 'example');

UPDATE products SET name = 'Trex' WHERE id = 1;
UPDATE products SET name = 'Edgen' WHERE id = 2;
UPDATE products SET name = 'example' WHERE id = 3;

-- Edgen 环境
INSERT OR IGNORE INTO environments (product_id, name, base_url, variables, is_default)
VALUES (2, 'dev', 'http://localhost:3001', '{}', 1);

INSERT OR IGNORE INTO environments (product_id, name, base_url, variables, is_default)
VALUES (2, 'staging', 'https://staging-edgen.example.com', '{}', 0);

-- example 环境
INSERT OR IGNORE INTO environments (product_id, name, base_url, variables, is_default)
VALUES (3, 'dev', 'http://localhost:3002', '{}', 1);
