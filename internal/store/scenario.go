package store

func (s *Store) CreateScenario(sc *Scenario) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO scenarios (product_id, folder_id, name, description, env_id)
		VALUES (?, ?, ?, ?, ?)`,
		sc.ProductID, sc.FolderID, sc.Name, sc.Description, sc.EnvID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) CreateScenarioStep(step *ScenarioStep) error {
	_, err := s.db.Exec(`
		INSERT INTO scenario_steps (scenario_id, step_order, name, api_id, method, path, headers, body, extract_rules, assertions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		step.ScenarioID, step.StepOrder, step.Name, step.APIID, step.Method, step.Path,
		step.Headers, step.Body, step.ExtractRules, step.Assertions)
	return err
}

func (s *Store) GetScenario(id int64) (*Scenario, error) {
	sc := &Scenario{}
	err := s.db.QueryRow(`
		SELECT id, product_id, folder_id, name, description, env_id, created_at
		FROM scenarios WHERE id = ?`, id).Scan(
		&sc.ID, &sc.ProductID, &sc.FolderID, &sc.Name, &sc.Description, &sc.EnvID, &sc.CreatedAt)
	if err != nil {
		return nil, err
	}
	if sc.FolderID > 0 {
		sc.FolderPath, _ = s.GetFolderPath(sc.FolderID)
	}
	steps, err := s.ListScenarioSteps(id)
	if err != nil {
		return nil, err
	}
	sc.Steps = steps
	return sc, nil
}

func (s *Store) ListScenarioSteps(scenarioID int64) ([]ScenarioStep, error) {
	rows, err := s.db.Query(`
		SELECT id, scenario_id, step_order, name, api_id, method, path, headers, body, extract_rules, assertions
		FROM scenario_steps WHERE scenario_id = ? ORDER BY step_order`, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ScenarioStep
	for rows.Next() {
		var st ScenarioStep
		if err := rows.Scan(&st.ID, &st.ScenarioID, &st.StepOrder, &st.Name, &st.APIID,
			&st.Method, &st.Path, &st.Headers, &st.Body, &st.ExtractRules, &st.Assertions); err != nil {
			return nil, err
		}
		list = append(list, st)
	}
	return list, rows.Err()
}

func (s *Store) ListScenarios(productID int64) ([]Scenario, error) {
	q := `
		SELECT id, product_id, folder_id, name, description, env_id, created_at
		FROM scenarios`
	args := []any{}
	if productID > 0 {
		q += ` WHERE product_id = ?`
		args = append(args, productID)
	}
	q += ` ORDER BY id DESC`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Scenario
	for rows.Next() {
		var sc Scenario
		if err := rows.Scan(&sc.ID, &sc.ProductID, &sc.FolderID, &sc.Name, &sc.Description, &sc.EnvID, &sc.CreatedAt); err != nil {
			return nil, err
		}
		if sc.FolderID > 0 {
			sc.FolderPath, _ = s.GetFolderPath(sc.FolderID)
		}
		list = append(list, sc)
	}
	return list, rows.Err()
}
