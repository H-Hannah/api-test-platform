-- 单文档验收规格（默认展示）；一键展开为 bdd-demo 三件套

ALTER TABLE bdd_features ADD COLUMN acceptance_md TEXT NOT NULL DEFAULT '';
