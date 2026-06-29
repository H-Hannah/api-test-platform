-- 环境 variables 仅保留五个服务 URL；token 改由测试数据管理
UPDATE environments
SET
  base_url = trim(
    coalesce(
      nullif(trim(json_extract(variables, '$.edgen_url')), ''),
      nullif(trim(json_extract(variables, '$.base_url_edgen')), ''),
      nullif(trim(json_extract(variables, '$.base_url')), ''),
      ''
    ),
    '/'
  ),
  variables = json_object(
    'edgen_url',
    trim(coalesce(
      nullif(trim(json_extract(variables, '$.edgen_url')), ''),
      nullif(trim(json_extract(variables, '$.base_url_edgen')), ''),
      ''
    )),
    'quest_edgen_url',
    trim(coalesce(nullif(trim(json_extract(variables, '$.quest_edgen_url')), ''), '')),
    'trex_url',
    trim(coalesce(
      nullif(trim(json_extract(variables, '$.trex_url')), ''),
      nullif(trim(json_extract(variables, '$.base_url_trex')), ''),
      ''
    )),
    'quest_trex_url',
    trim(coalesce(
      nullif(trim(json_extract(variables, '$.quest_trex_url')), ''),
      nullif(trim(json_extract(variables, '$.base_url_quest')), ''),
      ''
    )),
    'anchor_url',
    trim(coalesce(
      nullif(trim(json_extract(variables, '$.anchor_url')), ''),
      nullif(trim(json_extract(variables, '$.base_url_anchor')), ''),
      ''
    ))
  );
