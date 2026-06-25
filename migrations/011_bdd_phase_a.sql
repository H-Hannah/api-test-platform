-- 阶段 A：BDD 三件套（feature + testability-contract + traceability-matrix）+ Git 来源

ALTER TABLE bdd_features ADD COLUMN prd_git_path TEXT NOT NULL DEFAULT '';
ALTER TABLE bdd_features ADD COLUMN ui_git_path TEXT NOT NULL DEFAULT '';
ALTER TABLE bdd_features ADD COLUMN figma_url TEXT NOT NULL DEFAULT '';
ALTER TABLE bdd_features ADD COLUMN testability_contract TEXT NOT NULL DEFAULT '';
ALTER TABLE bdd_features ADD COLUMN traceability_matrix TEXT NOT NULL DEFAULT '';
ALTER TABLE bdd_features ADD COLUMN feature_files TEXT NOT NULL DEFAULT '[]';
ALTER TABLE bdd_features ADD COLUMN git_output_hint TEXT NOT NULL DEFAULT '';
ALTER TABLE bdd_features ADD COLUMN gate_passed INTEGER NOT NULL DEFAULT 0;
ALTER TABLE bdd_features ADD COLUMN gate_reasons TEXT NOT NULL DEFAULT '[]';
