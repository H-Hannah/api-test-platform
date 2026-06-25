-- 接口场景追溯：User Story、BDD、MR 标签（支持 MR 驱动测试）

ALTER TABLE api_definitions ADD COLUMN user_story TEXT NOT NULL DEFAULT '';
ALTER TABLE api_definitions ADD COLUMN bdd_ref TEXT NOT NULL DEFAULT '';
ALTER TABLE api_definitions ADD COLUMN mr_tags TEXT NOT NULL DEFAULT '';
