-- 阶段 C：接口场景追溯测试用例（TC），场景就绪以 TC 为准

ALTER TABLE api_definitions ADD COLUMN tc_ref TEXT NOT NULL DEFAULT '';
