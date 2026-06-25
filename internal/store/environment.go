package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// UpdateEnvironment 更新环境；若设为默认会取消其它环境的默认标记。
func (s *Store) UpdateEnvironment(e *Environment) error {
	if e.ID <= 0 {
		return fmt.Errorf("invalid environment id")
	}
	vars, err := NormalizeEnvVariablesJSON(e.BaseURL, e.Variables)
	if err != nil {
		return err
	}
	e.Variables = vars
	isDef := 0
	if e.IsDefault {
		isDef = 1
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if e.IsDefault {
		if _, err := tx.Exec(
			`UPDATE environments SET is_default = 0 WHERE id != ?`, e.ID); err != nil {
			return err
		}
	}
	res, err := tx.Exec(`
		UPDATE environments SET name = ?, base_url = ?, variables = ?, is_default = ?
		WHERE id = ?`,
		e.Name, e.BaseURL, e.Variables, isDef, e.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

// DeleteEnvironment 删除环境（至少保留一个环境时不允许删除最后一个）。
func (s *Store) DeleteEnvironment(id int64) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM environments`).Scan(&count); err != nil {
		return err
	}
	if count <= 1 {
		return fmt.Errorf("至少保留一个运行环境")
	}
	res, err := s.db.Exec(`DELETE FROM environments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// NormalizeEnvVariablesJSON 将 base_url 写入 variables.base_url，并校验 JSON 对象格式。
func NormalizeEnvVariablesJSON(baseURL, variables string) (string, error) {
	m := map[string]string{}
	raw := strings.TrimSpace(variables)
	if raw != "" && raw != "{}" {
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return "", fmt.Errorf("variables 必须是合法 JSON 对象")
		}
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base != "" {
		m["base_url"] = base
	}
	out, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ClearDefaultExcept 将除指定 id 外的环境设为非默认。
func (s *Store) ClearDefaultExcept(keepID int64) error {
	_, err := s.db.Exec(`UPDATE environments SET is_default = 0 WHERE id != ?`, keepID)
	return err
}
