-- 用例保存时记录接口定义指纹，用于检测过期（不依赖 meta 更新时间）

ALTER TABLE test_datasets ADD COLUMN api_fingerprint TEXT NOT NULL DEFAULT '';
