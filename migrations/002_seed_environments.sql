-- 环境种子数据（可重复执行）
-- 产品 Trex (id=1) 在 001 中已创建

-- Trex 环境
INSERT OR IGNORE INTO environments (product_id, name, base_url, variables, is_default)
VALUES (1, 'dev', 'http://localhost:3000', '{"username":"tester","password":"test123","token":""}', 0);

-- T-Rex 联调环境（前端 https://www.trex.xyz ，API 按实际抓包修改 base_url）
INSERT OR IGNORE INTO environments (product_id, name, base_url, variables, is_default)
VALUES (1, 'trex-dev', 'https://api.trex.xyz', '{"token":"","walletAddress":"0x0000000000000000000000000000000000000000"}', 1);

INSERT OR IGNORE INTO environments (product_id, name, base_url, variables, is_default)
VALUES (1, 'staging', 'https://staging-a.example.com', '{"username":"staging_user","password":"staging_pass"}', 0);
