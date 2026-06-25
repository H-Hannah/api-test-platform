package store

import (
	"encoding/json"
	"time"
)

// BDDFeatureFile 单个 .feature 文件。
type BDDFeatureFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// BDDFeature 阶段 A：PRD + UI 设计生成的 BDD 三件套（平台暂存，人工提交 Git）。
type BDDFeature struct {
	ID                   int64            `json:"id"`
	ProductID            int64            `json:"product_id"`
	Title                string           `json:"title"`
	UserStory            string           `json:"user_story"`
	PRDGitPath           string           `json:"prd_git_path"`
	UIGitPath            string           `json:"ui_git_path"`
	FigmaURL             string           `json:"figma_url"`
	PRDText              string           `json:"prd_text"`
	UIDesignText         string           `json:"ui_design_text"`
	Gherkin              string           `json:"gherkin"`
	TestabilityContract  string           `json:"testability_contract"`
	TraceabilityMatrix   string           `json:"traceability_matrix"`
	FeatureFiles         []BDDFeatureFile `json:"feature_files"`
	AcceptanceMD         string           `json:"acceptance_md"`
	GitOutputHint        string           `json:"git_output_hint"`
	GatePassed           bool             `json:"gate_passed"`
	GateReasons          []string         `json:"gate_reasons"`
	CreatedAt            string           `json:"created_at"`
	UpdatedAt            string           `json:"updated_at"`
}

const bddSelectCols = `id, product_id, title, user_story, prd_git_path, ui_git_path, figma_url,
	prd_text, design_text, gherkin, testability_contract, traceability_matrix, feature_files,
	acceptance_md, git_output_hint, gate_passed, gate_reasons, created_at, updated_at`

func scanBDDFeature(row interface{ Scan(...any) error }) (*BDDFeature, error) {
	var f BDDFeature
	var featureFilesJSON, gateReasonsJSON string
	var gatePassed int
	var designText string
	err := row.Scan(
		&f.ID, &f.ProductID, &f.Title, &f.UserStory, &f.PRDGitPath, &f.UIGitPath, &f.FigmaURL,
		&f.PRDText, &designText, &f.Gherkin, &f.TestabilityContract, &f.TraceabilityMatrix, &featureFilesJSON,
		&f.AcceptanceMD, &f.GitOutputHint, &gatePassed, &gateReasonsJSON, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	f.UIDesignText = designText
	f.GatePassed = gatePassed != 0
	_ = json.Unmarshal([]byte(featureFilesJSON), &f.FeatureFiles)
	if f.FeatureFiles == nil {
		f.FeatureFiles = []BDDFeatureFile{}
	}
	_ = json.Unmarshal([]byte(gateReasonsJSON), &f.GateReasons)
	if f.GateReasons == nil {
		f.GateReasons = []string{}
	}
	return &f, nil
}

func (s *Store) CreateBDDFeature(f *BDDFeature) (int64, error) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	ff, gr := bddJSONFields(f)
	res, err := s.db.Exec(`
		INSERT INTO bdd_features (
			product_id, title, user_story, prd_git_path, ui_git_path, figma_url,
			prd_text, design_text, gherkin, testability_contract, traceability_matrix,
			feature_files, acceptance_md, git_output_hint, gate_passed, gate_reasons, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ProductID, f.Title, f.UserStory, f.PRDGitPath, f.UIGitPath, f.FigmaURL,
		f.PRDText, f.UIDesignText, f.Gherkin, f.TestabilityContract, f.TraceabilityMatrix,
		ff, f.AcceptanceMD, f.GitOutputHint, boolToInt(f.GatePassed), gr, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetBDDFeature(id int64) (*BDDFeature, error) {
	row := s.db.QueryRow(`SELECT `+bddSelectCols+` FROM bdd_features WHERE id = ?`, id)
	return scanBDDFeature(row)
}

func (s *Store) ListBDDFeatures(productID int64) ([]BDDFeature, error) {
	rows, err := s.db.Query(`SELECT `+bddSelectCols+` FROM bdd_features WHERE product_id = ? ORDER BY updated_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BDDFeature
	for rows.Next() {
		f, err := scanBDDFeature(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *f)
	}
	return list, rows.Err()
}

func (s *Store) DeleteBDDFeature(id, productID int64) error {
	_, err := s.db.Exec(`DELETE FROM bdd_features WHERE id = ? AND product_id = ?`, id, productID)
	return err
}

func (s *Store) UpdateBDDFeature(f *BDDFeature) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	ff, gr := bddJSONFields(f)
	_, err := s.db.Exec(`
		UPDATE bdd_features SET
			title = ?, user_story = ?, gherkin = ?, testability_contract = ?, traceability_matrix = ?,
			feature_files = ?, acceptance_md = ?, git_output_hint = ?, gate_passed = ?, gate_reasons = ?,
			updated_at = ?
		WHERE id = ? AND product_id = ?`,
		f.Title, f.UserStory, f.Gherkin, f.TestabilityContract, f.TraceabilityMatrix,
		ff, f.AcceptanceMD, f.GitOutputHint, boolToInt(f.GatePassed), gr, now, f.ID, f.ProductID)
	return err
}

func bddJSONFields(f *BDDFeature) (featureFiles string, gateReasons string) {
	ff, _ := json.Marshal(f.FeatureFiles)
	if len(f.FeatureFiles) == 0 {
		ff = []byte("[]")
	}
	gr, _ := json.Marshal(f.GateReasons)
	if len(f.GateReasons) == 0 {
		gr = []byte("[]")
	}
	return string(ff), string(gr)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
