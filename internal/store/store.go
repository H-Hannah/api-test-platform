package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) Migrate(sqlFile string) error {
	b, err := os.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	base := filepath.Base(sqlFile)
	for i, stmt := range splitMigrationStatements(string(b)) {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration %s stmt %d: %w", base, i+1, err)
		}
	}
	return nil
}

// splitMigrationStatements 按分号拆分 SQL（忽略行注释），逐条执行。
func splitMigrationStatements(sql string) []string {
	var out []string
	var b strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	raw := b.String()
	start := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] != ';' {
			continue
		}
		stmt := strings.TrimSpace(raw[start : i+1])
		stmt = strings.TrimSuffix(stmt, ";")
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			out = append(out, stmt)
		}
		start = i + 1
	}
	tail := strings.TrimSpace(raw[start:])
	if tail != "" {
		out = append(out, tail)
	}
	return out
}
