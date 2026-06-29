package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type TestDatasetFilter struct {
	ProductID     int64
	Version       string
	RequirementID string
	APIID         int64 // 按 method+path 匹配 api_bindings
}

func (s *Store) CreateTestDataset(ds *TestDataset) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO test_datasets (
			product_id, version, requirement_id, dataset_key, name, description,
			tc_refs, api_bindings, variables, headers_override, body_override,
			obtain_type, obtain_note, owner, tags, source, assertions, api_fingerprint
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ds.ProductID, ds.Version, ds.RequirementID, ds.DatasetKey, ds.Name, ds.Description,
		ds.TcRefs, ds.ApiBindings, ds.Variables, ds.HeadersOverride, ds.BodyOverride,
		ds.ObtainType, ds.ObtainNote, ds.Owner, ds.Tags, ds.Source, defaultDatasetAssertions(ds.Assertions),
		strings.TrimSpace(ds.ApiFingerprint))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpsertTestDataset(ds *TestDataset) (int64, error) {
	key := strings.TrimSpace(ds.DatasetKey)
	if key == "" {
		return s.CreateTestDataset(ds)
	}
	var existing int64
	err := s.db.QueryRow(`
		SELECT id FROM test_datasets
		WHERE product_id = ? AND version = ? AND requirement_id = ? AND dataset_key = ?`,
		ds.ProductID, ds.Version, ds.RequirementID, key).Scan(&existing)
	if err == sql.ErrNoRows {
		return s.CreateTestDataset(ds)
	}
	if err != nil {
		return 0, err
	}
	_, err = s.db.Exec(`
		UPDATE test_datasets SET
			name = ?, description = ?, tc_refs = ?, api_bindings = ?,
			variables = ?, headers_override = ?, body_override = ?,
			obtain_type = ?, obtain_note = ?, owner = ?, tags = ?, source = ?,
			assertions = ?, api_fingerprint = ?, updated_at = datetime('now')
		WHERE id = ?`,
		ds.Name, ds.Description, ds.TcRefs, ds.ApiBindings,
		ds.Variables, ds.HeadersOverride, ds.BodyOverride,
		ds.ObtainType, ds.ObtainNote, ds.Owner, ds.Tags, ds.Source,
		defaultDatasetAssertions(ds.Assertions), strings.TrimSpace(ds.ApiFingerprint), existing)
	if err != nil {
		return 0, err
	}
	return existing, nil
}

func (s *Store) GetTestDataset(id int64) (*TestDataset, error) {
	row := s.db.QueryRow(`
		SELECT id, product_id, version, requirement_id, dataset_key, name, description,
			tc_refs, api_bindings, variables, headers_override, body_override,
			obtain_type, obtain_note, owner, tags, source, assertions, api_fingerprint, created_at, updated_at
		FROM test_datasets WHERE id = ?`, id)
	return scanTestDataset(row)
}

func (s *Store) ListTestDatasets(f TestDatasetFilter) ([]TestDataset, error) {
	q := `
		SELECT id, product_id, version, requirement_id, dataset_key, name, description,
			tc_refs, api_bindings, variables, headers_override, body_override,
			obtain_type, obtain_note, owner, tags, source, assertions, api_fingerprint, created_at, updated_at
		FROM test_datasets WHERE 1=1`
	args := []any{}
	if f.ProductID > 0 {
		q += ` AND product_id = ?`
		args = append(args, f.ProductID)
	}
	if v := strings.TrimSpace(f.Version); v != "" {
		q += ` AND version = ?`
		args = append(args, v)
	}
	if rid := strings.TrimSpace(f.RequirementID); rid != "" {
		q += ` AND requirement_id = ?`
		args = append(args, rid)
	}
	q += ` ORDER BY dataset_key, id`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TestDataset
	for rows.Next() {
		ds, err := scanTestDatasetRows(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *ds)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if f.APIID <= 0 {
		return list, nil
	}
	api, err := s.getAPIRow(f.APIID)
	if err != nil {
		return nil, err
	}
	var matched []TestDataset
	for _, ds := range list {
		if DatasetBelongsToAPI(&ds, api) {
			matched = append(matched, ds)
		}
	}
	return matched, nil
}

func (s *Store) CountAPICases(apiID int64) (int, error) {
	list, err := s.ListTestDatasets(TestDatasetFilter{APIID: apiID})
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

func (s *Store) DeleteTestDataset(id int64) error {
	_, err := s.db.Exec(`DELETE FROM test_datasets WHERE id = ?`, id)
	return err
}

// UpdateTestDataset 更新可编辑字段（名称、请求覆盖、断言等）。
func (s *Store) UpdateTestDataset(ds *TestDataset) error {
	if ds.ID <= 0 {
		return fmt.Errorf("invalid dataset id")
	}
	res, err := s.db.Exec(`
		UPDATE test_datasets SET
			name = ?, description = ?, variables = ?, headers_override = ?,
			body_override = ?, assertions = ?, tags = ?, api_fingerprint = ?, updated_at = datetime('now')
		WHERE id = ?`,
		ds.Name, ds.Description, ds.Variables, ds.HeadersOverride, ds.BodyOverride,
		defaultDatasetAssertions(ds.Assertions), defaultDatasetTags(ds.Tags), strings.TrimSpace(ds.ApiFingerprint), ds.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) UpsertTestDataSpec(spec *TestDataSpec) (int64, error) {
	var existing int64
	err := s.db.QueryRow(`
		SELECT id FROM test_data_specs
		WHERE product_id = ? AND version = ? AND requirement_id = ?`,
		spec.ProductID, spec.Version, spec.RequirementID).Scan(&existing)
	if err == sql.ErrNoRows {
		res, err := s.db.Exec(`
			INSERT INTO test_data_specs (product_id, version, requirement_id, requirement_name, spec_yaml, env_keys)
			VALUES (?, ?, ?, ?, ?, ?)`,
			spec.ProductID, spec.Version, spec.RequirementID, spec.RequirementName, spec.SpecYAML, spec.EnvKeys)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	if err != nil {
		return 0, err
	}
	_, err = s.db.Exec(`
		UPDATE test_data_specs SET requirement_name = ?, spec_yaml = ?, env_keys = ?, updated_at = datetime('now')
		WHERE id = ?`,
		spec.RequirementName, spec.SpecYAML, spec.EnvKeys, existing)
	return existing, err
}

func (s *Store) GetTestDataSpec(productID int64, version, requirementID string) (*TestDataSpec, error) {
	row := s.db.QueryRow(`
		SELECT id, product_id, version, requirement_id, requirement_name, spec_yaml, env_keys, created_at, updated_at
		FROM test_data_specs
		WHERE product_id = ? AND version = ? AND requirement_id = ?`,
		productID, version, requirementID)
	var spec TestDataSpec
	err := row.Scan(&spec.ID, &spec.ProductID, &spec.Version, &spec.RequirementID,
		&spec.RequirementName, &spec.SpecYAML, &spec.EnvKeys, &spec.CreatedAt, &spec.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("test data spec not found")
	}
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (s *Store) MergeEnvVarKeys(envID int64, keys []string, placeholder string) (int, error) {
	env, err := s.GetEnvironment(envID)
	if err != nil {
		return 0, err
	}
	vars := map[string]string{}
	_ = json.Unmarshal([]byte(env.Variables), &vars)
	if vars == nil {
		vars = map[string]string{}
	}
	if placeholder == "" {
		placeholder = ""
	}
	added := 0
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := vars[k]; ok {
			continue
		}
		vars[k] = placeholder
		added++
	}
	b, _ := json.Marshal(vars)
	if err := s.UpdateEnvironmentVariables(envID, string(b)); err != nil {
		return 0, err
	}
	return added, nil
}

func (s *Store) UpdateEnvironmentVariables(envID int64, variables string) error {
	_, err := s.db.Exec(`UPDATE environments SET variables = ? WHERE id = ?`, variables, envID)
	return err
}

func defaultDatasetAssertions(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "[]"
	}
	return raw
}

func defaultDatasetTags(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "[]"
	}
	return raw
}

// EnrichDatasetStale 当接口定义指纹与用例保存时不一致时标记过期。
func (s *Store) EnrichDatasetStale(apiID int64, ds *TestDataset) {
	if ds == nil || apiID <= 0 {
		return
	}
	api, err := s.getAPIRow(apiID)
	if err != nil {
		return
	}
	if stale, reason := DatasetStaleAgainstAPI(api, ds); stale {
		ds.Stale = true
		ds.StaleReason = reason
	}
}

func datasetStaleAgainstAPI(apiUpdated, dsUpdated string) bool {
	a := normalizeSQLiteTime(apiUpdated)
	b := normalizeSQLiteTime(dsUpdated)
	if a == "" || b == "" {
		return false
	}
	return a > b
}

func normalizeSQLiteTime(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	s = strings.Replace(s, "T", " ", 1)
	if len(s) >= 19 {
		return s[:19]
	}
	return s
}

func scanTestDataset(row *sql.Row) (*TestDataset, error) {
	var ds TestDataset
	err := row.Scan(&ds.ID, &ds.ProductID, &ds.Version, &ds.RequirementID, &ds.DatasetKey,
		&ds.Name, &ds.Description, &ds.TcRefs, &ds.ApiBindings, &ds.Variables,
		&ds.HeadersOverride, &ds.BodyOverride, &ds.ObtainType, &ds.ObtainNote,
		&ds.Owner, &ds.Tags, &ds.Source, &ds.Assertions, &ds.ApiFingerprint, &ds.CreatedAt, &ds.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func scanTestDatasetRows(rows *sql.Rows) (*TestDataset, error) {
	var ds TestDataset
	err := rows.Scan(&ds.ID, &ds.ProductID, &ds.Version, &ds.RequirementID, &ds.DatasetKey,
		&ds.Name, &ds.Description, &ds.TcRefs, &ds.ApiBindings, &ds.Variables,
		&ds.HeadersOverride, &ds.BodyOverride, &ds.ObtainType, &ds.ObtainNote,
		&ds.Owner, &ds.Tags, &ds.Source, &ds.Assertions, &ds.ApiFingerprint, &ds.CreatedAt, &ds.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}
