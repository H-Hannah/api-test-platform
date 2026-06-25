-- 每个部署环境（BETA/PRE/PROD）配置多微服务 base URL，场景执行时一次选环境即可跨 trex/quest/anchor

-- 若库中仍是旧环境名，先升级到 BETA/PRE/PROD（与 005 一致，可重复执行）
UPDATE scenarios SET env_id = NULL WHERE env_id IS NOT NULL;
DELETE FROM run_steps;
DELETE FROM runs;
DELETE FROM environments;

INSERT INTO environments (product_id, name, base_url, variables, is_default) VALUES
(1, 'BETA', 'https://api.trex.beta.dipbit.xyz',
 '{"base_url":"https://api.trex.beta.dipbit.xyz","base_url_trex":"https://api.trex.beta.dipbit.xyz","base_url_quest":"https://api.quests.beta.dipbit.xyz","base_url_anchor":"https://anchor.trex.beta.dipbit.xyz"}', 0),
(1, 'PRE', 'https://api.trex.pre.dipbit.xyz',
 '{"base_url":"https://api.trex.pre.dipbit.xyz","base_url_trex":"https://api.trex.pre.dipbit.xyz","base_url_quest":"https://api.quests.pre.dipbit.xyz","base_url_anchor":"https://anchor.trex.pre.dipbit.xyz"}', 0),
(1, 'PROD', 'https://api.trex.xyz',
 '{"base_url":"https://api.trex.xyz","base_url_trex":"https://api.trex.xyz","base_url_quest":"https://quest.trex.xyz","base_url_anchor":"https://anchor.trex.xyz"}', 1),
(2, 'BETA', 'https://api.beta.ospprotocol.xyz',
 '{"base_url":"https://api.beta.ospprotocol.xyz","base_url_edgen":"https://api.beta.ospprotocol.xyz","base_url_quest":"https://api.quests.beta.dipbit.xyz","base_url_openreplay":"https://openreplay.ospprotocol.xyz"}', 0),
(2, 'PRE', 'https://api.pre.ospprotocol.xyz',
 '{"base_url":"https://api.pre.ospprotocol.xyz","base_url_edgen":"https://api.pre.ospprotocol.xyz","base_url_quest":"https://api.quests.pre.dipbit.xyz","base_url_openreplay":"https://openreplay.pre.ospprotocol.xyz"}', 0),
(2, 'PROD', 'https://api.edgen.tech',
 '{"base_url":"https://api.edgen.tech","base_url_edgen":"https://api.edgen.tech","base_url_quest":"https://quest.edgen.tech","base_url_openreplay":"https://openreplay.edgen.tech"}', 1),
(3, 'BETA', 'http://localhost:3002',
 '{"base_url":"http://localhost:3002","base_url_example":"http://localhost:3002"}', 1),
(3, 'PRE', 'http://pre-api.example.com',
 '{"base_url":"http://pre-api.example.com","base_url_example":"http://pre-api.example.com"}', 0),
(3, 'PROD', 'http://prod-api.example.com',
 '{"base_url":"http://prod-api.example.com","base_url_example":"http://prod-api.example.com"}', 0);
