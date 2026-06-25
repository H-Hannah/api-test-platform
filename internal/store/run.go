package store

func (s *Store) CreateRun(r *Run) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO runs (scenario_id, api_id, env_id, status, summary) VALUES (?, ?, ?, ?, ?)`,
		r.ScenarioID, r.APIID, r.EnvID, r.Status, r.Summary)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) FinishRun(id int64, status, summary string) error {
	_, err := s.db.Exec(`
		UPDATE runs SET status = ?, summary = ?, finished_at = datetime('now') WHERE id = ?`,
		status, summary, id)
	return err
}

func (s *Store) CreateRunStep(step *RunStep) (int64, error) {
	res, err := s.db.Exec(`
		INSERT INTO run_steps (run_id, step_order, name, status, request_snapshot, response_snapshot,
			assertion_results, duration_ms, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		step.RunID, step.StepOrder, step.Name, step.Status, step.RequestSnapshot,
		step.ResponseSnapshot, step.AssertionResults, step.DurationMS, step.ErrorMessage)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetRun(id int64) (*Run, error) {
	r := &Run{}
	err := s.db.QueryRow(`
		SELECT id, scenario_id, api_id, env_id, status, started_at, finished_at, summary
		FROM runs WHERE id = ?`, id).Scan(
		&r.ID, &r.ScenarioID, &r.APIID, &r.EnvID, &r.Status, &r.StartedAt, &r.FinishedAt, &r.Summary)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`
		SELECT id, run_id, step_order, name, status, request_snapshot, response_snapshot,
			assertion_results, duration_ms, error_message
		FROM run_steps WHERE run_id = ? ORDER BY step_order`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var st RunStep
		if err := rows.Scan(&st.ID, &st.RunID, &st.StepOrder, &st.Name, &st.Status,
			&st.RequestSnapshot, &st.ResponseSnapshot, &st.AssertionResults,
			&st.DurationMS, &st.ErrorMessage); err != nil {
			return nil, err
		}
		r.Steps = append(r.Steps, st)
	}
	return r, rows.Err()
}

func (s *Store) ListRunsByAPI(apiID int64, limit int) ([]Run, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`
		SELECT id, scenario_id, api_id, env_id, status, started_at, finished_at, summary
		FROM runs WHERE api_id = ? ORDER BY id DESC LIMIT ?`, apiID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.ScenarioID, &r.APIID, &r.EnvID, &r.Status,
			&r.StartedAt, &r.FinishedAt, &r.Summary); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Store) ListRuns(limit int) ([]Run, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, scenario_id, api_id, env_id, status, started_at, finished_at, summary
		FROM runs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.ScenarioID, &r.APIID, &r.EnvID, &r.Status,
			&r.StartedAt, &r.FinishedAt, &r.Summary); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}
