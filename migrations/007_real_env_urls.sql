-- 真实抓包域名（用户提供）；仅 UPDATE，不重建环境表

-- Trex (product_id=1)
UPDATE environments SET
  base_url = 'https://api.trex.beta.dipbit.xyz',
  variables = '{"base_url":"https://api.trex.beta.dipbit.xyz","base_url_trex":"https://api.trex.beta.dipbit.xyz","base_url_quest":"https://api.quests.beta.dipbit.xyz","base_url_anchor":"https://anchor.trex.beta.dipbit.xyz"}'
WHERE product_id = 1 AND name = 'BETA';

UPDATE environments SET
  base_url = 'https://api.trex.xyz',
  variables = '{"base_url":"https://api.trex.xyz","base_url_trex":"https://api.trex.xyz","base_url_quest":"https://quest.trex.xyz","base_url_anchor":"https://anchor.trex.xyz"}'
WHERE product_id = 1 AND name = 'PROD';

-- Edgen (product_id=2)
UPDATE environments SET
  base_url = 'https://api.beta.ospprotocol.xyz',
  variables = '{"base_url":"https://api.beta.ospprotocol.xyz","base_url_edgen":"https://api.beta.ospprotocol.xyz","base_url_quest":"https://api.quests.beta.dipbit.xyz","base_url_openreplay":"https://openreplay.ospprotocol.xyz"}'
WHERE product_id = 2 AND name = 'BETA';

UPDATE environments SET
  base_url = 'https://api.edgen.tech',
  variables = '{"base_url":"https://api.edgen.tech","base_url_edgen":"https://api.edgen.tech","base_url_quest":"https://quest.edgen.tech","base_url_openreplay":"https://openreplay.edgen.tech"}'
WHERE product_id = 2 AND name = 'PROD';
