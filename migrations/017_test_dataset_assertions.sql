-- 测试数据集支持每条用例自带断言（单接口多案例）

ALTER TABLE test_datasets ADD COLUMN assertions TEXT NOT NULL DEFAULT '[]';
