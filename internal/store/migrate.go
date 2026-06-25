package store

import "path/filepath"

var migrationFiles = []string{
	"001_init.sql",
	"002_seed_environments.sql",
	"003_product_trex.sql",
	"004_products_trex_edgen_example.sql",
	"005_products_envs_beta_pre_prod.sql",
	"006_env_multi_service_variables.sql",
	"007_real_env_urls.sql",
	"008_env_variables_token.sql",
	"009_api_test_metadata.sql",
	"010_bdd_features.sql",
	"011_bdd_phase_a.sql",
	"012_api_tc_ref.sql",
	"013_bdd_acceptance_md.sql",
	"014_test_datasets.sql",
	"015_global_environments.sql",
}

func (s *Store) ensureMigrationTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	return err
}

func (s *Store) migrationApplied(name string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE name = ?`, name).Scan(&n)
	return n > 0, err
}

func (s *Store) markMigrationApplied(name string) error {
	_, err := s.db.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name)
	return err
}

// bootstrapMigrationsIfNeeded 兼容已有库：在引入 schema_migrations 之前已跑过迁移时，跳过 001–005 仅执行新迁移。
func (s *Store) bootstrapMigrationsIfNeeded() error {
	var tracked int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM schema_migrations`).Scan(&tracked); err != nil {
		return err
	}
	if tracked > 0 {
		return nil
	}
	var envCount int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM environments`).Scan(&envCount); err != nil {
		return err
	}
	if envCount == 0 {
		return nil
	}
	var modern int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM environments WHERE name IN ('BETA','PRE','PROD')`).Scan(&modern); err != nil {
		return err
	}
	// 已是 BETA/PRE/PROD：只补跑 006
	if modern > 0 {
		for _, name := range migrationFiles {
			if name == "006_env_multi_service_variables.sql" {
				break
			}
			if err := s.markMigrationApplied(name); err != nil {
				return err
			}
		}
		return nil
	}
	// 仍是 dev/staging 等旧环境：从 005 重跑
	for _, name := range migrationFiles {
		if name == "005_products_envs_beta_pre_prod.sql" {
			break
		}
		if err := s.markMigrationApplied(name); err != nil {
			return err
		}
	}
	return nil
}

// MigrateAll runs SQL migrations in order (each file applied at most once).
func (s *Store) MigrateAll(dir string) error {
	if err := s.ensureMigrationTable(); err != nil {
		return err
	}
	if err := s.bootstrapMigrationsIfNeeded(); err != nil {
		return err
	}
	for _, name := range migrationFiles {
		applied, err := s.migrationApplied(name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := s.Migrate(filepath.Join(dir, name)); err != nil {
			return err
		}
		if err := s.markMigrationApplied(name); err != nil {
			return err
		}
	}
	return nil
}
