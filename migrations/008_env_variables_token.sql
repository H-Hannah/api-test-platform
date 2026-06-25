-- 为各环境 variables 补充 token 占位（执行时替换 Authorization 中的 {{token}}）
-- 请在 Web/DB 中填入真实 token，或保持空字符串仅调试无需鉴权的接口

UPDATE environments SET variables = '{"base_url":"https://api.trex.beta.dipbit.xyz","base_url_trex":"https://api.trex.beta.dipbit.xyz","base_url_quest":"https://api.quests.beta.dipbit.xyz","base_url_anchor":"https://anchor.trex.beta.dipbit.xyz","token":""}'
WHERE product_id = 1 AND name = 'BETA';

UPDATE environments SET variables = '{"base_url":"https://api.trex.xyz","base_url_trex":"https://api.trex.xyz","base_url_quest":"https://quest.trex.xyz","base_url_anchor":"https://anchor.trex.xyz","token":""}'
WHERE product_id = 1 AND name = 'PROD';

UPDATE environments SET variables = '{"base_url":"https://api.beta.ospprotocol.xyz","base_url_edgen":"https://api.beta.ospprotocol.xyz","base_url_quest":"https://api.quests.beta.dipbit.xyz","base_url_openreplay":"https://openreplay.ospprotocol.xyz","token":""}'
WHERE product_id = 2 AND name = 'BETA';

UPDATE environments SET variables = '{"base_url":"https://api.edgen.tech","base_url_edgen":"https://api.edgen.tech","base_url_quest":"https://quest.edgen.tech","base_url_openreplay":"https://openreplay.edgen.tech","token":""}'
WHERE product_id = 2 AND name = 'PROD';
