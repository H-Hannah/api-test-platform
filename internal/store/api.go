package store

import (
	"database/sql"
	"encoding/json"
)

func (s *Store) CreateAPI(api *APIDefinition) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO api_definitions (
			product_id, folder_id, name, method, path, full_url_template,
			headers, body, body_type, description, ai_remark, source_record
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		api.ProductID, api.FolderID, api.Name, api.Method, api.Path, api.FullURLTemplate,
		api.Headers, api.Body, api.BodyType, api.Description, api.AIRemark, api.SourceRecord)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) CreateAssertions(apiID int64, assertions []Assertion) error {
	for _, a := range assertions {
		enabled := 1
		if !a.Enabled && a.APIID != 0 {
			enabled = 0
		}
		_, err := s.db.Exec(`
			INSERT INTO api_assertions (api_id, type, expression, operator, expected, enabled)
			VALUES (?, ?, ?, ?, ?, ?)`,
			apiID, a.Type, a.Expression, a.Operator, a.Expected, enabled)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetAPI(id int64) (*APIDefinition, error) {
	api, err := s.getAPIRow(id)
	if err != nil {
		return nil, err
	}
	if api.FolderID > 0 {
		api.FolderPath, _ = s.GetFolderPath(api.FolderID)
	}
	assertions, err := s.ListAssertions(id)
	if err != nil {
		return nil, err
	}
	api.Assertions = assertions
	api.AssertionCount = len(assertions)
	cc, _ := s.CountAPICases(id)
	fillAPIScenarioFlags(api, cc)
	return api, nil
}

func (s *Store) ListAssertions(apiID int64) ([]Assertion, error) {
	rows, err := s.db.Query(`
		SELECT id, api_id, type, expression, operator, expected, enabled
		FROM api_assertions WHERE api_id = ?`, apiID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Assertion
	for rows.Next() {
		var a Assertion
		var enabled int
		if err := rows.Scan(&a.ID, &a.APIID, &a.Type, &a.Expression, &a.Operator, &a.Expected, &enabled); err != nil {
			return nil, err
		}
		a.Enabled = enabled == 1
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) ListAPIs(productID int64, folderID int64) ([]APIDefinition, error) {
	return s.ListAPIsFiltered(APIListFilter{ProductID: productID, FolderID: folderID})
}

func (s *Store) DeleteAPI(id int64) error {
	reqID := APIRequirementID(id)
	if _, err := s.db.Exec(`DELETE FROM test_datasets WHERE requirement_id = ?`, reqID); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM api_assertions WHERE api_id = ?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM api_definitions WHERE id = ?`, id)
	return err
}

func (s *Store) ListProducts() ([]Product, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Store) GetEnvironment(id int64) (*Environment, error) {
	e := &Environment{}
	var isDef int
	err := s.db.QueryRow(`
		SELECT id, name, base_url, variables, is_default, created_at
		FROM environments WHERE id = ?`, id).Scan(
		&e.ID, &e.Name, &e.BaseURL, &e.Variables, &isDef, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	e.IsDefault = isDef == 1
	return e, nil
}

func (s *Store) ListEnvironments(_ int64) ([]Environment, error) {
	return s.ListAllEnvironments()
}

func (s *Store) ListAllEnvironments() ([]Environment, error) {
	rows, err := s.db.Query(`
		SELECT id, name, base_url, variables, is_default, created_at
		FROM environments ORDER BY is_default DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Environment
	for rows.Next() {
		var e Environment
		var isDef int
		if err := rows.Scan(&e.ID, &e.Name, &e.BaseURL, &e.Variables, &isDef, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.IsDefault = isDef == 1
		list = append(list, e)
	}
	return list, rows.Err()
}

func (s *Store) CreateEnvironment(e *Environment) (int64, error) {
	isDef := 0
	if e.IsDefault {
		isDef = 1
	}
	res, err := s.db.Exec(`
		INSERT INTO environments (name, base_url, variables, is_default)
		VALUES (?, ?, ?, ?)`,
		e.Name, e.BaseURL, e.Variables, isDef)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) SourceRecordJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *Store) DefaultEnv() (*Environment, error) {
	e := &Environment{}
	var isDef int
	err := s.db.QueryRow(`
		SELECT id, name, base_url, variables, is_default, created_at
		FROM environments
		ORDER BY is_default DESC, id LIMIT 1`).Scan(
		&e.ID, &e.Name, &e.BaseURL, &e.Variables, &isDef, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.IsDefault = isDef == 1
	return e, nil
}

// DefaultEnvForProduct 兼容旧调用，环境已全局化。
func (s *Store) DefaultEnvForProduct(_ int64) (*Environment, error) {
	return s.DefaultEnv()
}
