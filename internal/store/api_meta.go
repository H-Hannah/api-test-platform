package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// APIListFilter 列表筛选（MR / US / 缺口）。
type APIListFilter struct {
	ProductID  int64
	FolderID   int64
	UserStory  string
	MRTag      string
	Gap        string // no_us | no_bdd | no_tc | no_assert | not_ready | ready
}

const apiSelectCols = `a.id, a.product_id, a.folder_id, a.name, a.method, a.path, a.full_url_template,
	a.headers, a.body, a.body_type, a.description, a.ai_remark, a.source_record,
	a.user_story, a.bdd_ref, a.tc_ref, a.mr_tags, a.created_at, a.updated_at`

func scanAPI(row interface {
	Scan(dest ...any) error
}, api *APIDefinition, assertionCount *int) error {
	dest := []any{
		&api.ID, &api.ProductID, &api.FolderID, &api.Name, &api.Method, &api.Path,
		&api.FullURLTemplate, &api.Headers, &api.Body, &api.BodyType,
		&api.Description, &api.AIRemark, &api.SourceRecord,
		&api.UserStory, &api.BDDRef, &api.TCRef, &api.MRTags, &api.CreatedAt, &api.UpdatedAt,
	}
	if assertionCount != nil {
		dest = append(dest, assertionCount)
	}
	return row.Scan(dest...)
}

func fillAPIScenarioFlags(api *APIDefinition, assertionCount int) {
	api.AssertionCount = assertionCount
	api.ScenarioReady = strings.TrimSpace(api.UserStory) != "" &&
		strings.TrimSpace(api.TCRef) != "" &&
		assertionCount > 0
}

func (s *Store) ListAPIsFiltered(f APIListFilter) ([]APIDefinition, error) {
	q := `SELECT ` + apiSelectCols + `,
		(SELECT COUNT(1) FROM api_assertions WHERE api_id = a.id AND enabled = 1) AS ac
		FROM api_definitions a WHERE 1=1`
	args := []any{}
	if f.ProductID > 0 {
		q += ` AND a.product_id = ?`
		args = append(args, f.ProductID)
	}
	if f.FolderID > 0 {
		q += ` AND a.folder_id = ?`
		args = append(args, f.FolderID)
	}
	if us := strings.TrimSpace(f.UserStory); us != "" {
		q += ` AND a.user_story LIKE ?`
		args = append(args, "%"+us+"%")
	}
	if mr := strings.TrimSpace(f.MRTag); mr != "" {
		q += ` AND (',' || a.mr_tags || ',') LIKE ?`
		args = append(args, "%,"+mr+",%")
	}
	switch strings.TrimSpace(f.Gap) {
	case "no_us", "no_user_story":
		q += ` AND TRIM(a.user_story) = ''`
	case "no_bdd":
		q += ` AND TRIM(a.bdd_ref) = ''`
	case "no_tc":
		q += ` AND TRIM(a.tc_ref) = ''`
	case "no_assert", "no_assertions":
		q += ` AND (SELECT COUNT(1) FROM api_assertions WHERE api_id = a.id AND enabled = 1) = 0`
	case "not_ready":
		q += ` AND NOT (
			TRIM(a.user_story) != '' AND TRIM(a.tc_ref) != '' AND
			(SELECT COUNT(1) FROM api_assertions WHERE api_id = a.id AND enabled = 1) > 0
		)`
	case "ready":
		q += ` AND TRIM(a.user_story) != '' AND TRIM(a.tc_ref) != '' AND
			(SELECT COUNT(1) FROM api_assertions WHERE api_id = a.id AND enabled = 1) > 0`
	}
	q += ` ORDER BY a.updated_at DESC`
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []APIDefinition
	for rows.Next() {
		var api APIDefinition
		var ac int
		if err := scanAPI(rows, &api, &ac); err != nil {
			return nil, err
		}
		if api.FolderID > 0 {
			api.FolderPath, _ = s.GetFolderPath(api.FolderID)
		}
		fillAPIScenarioFlags(&api, ac)
		list = append(list, api)
	}
	return list, rows.Err()
}

func (s *Store) UpdateAPIMeta(id int64, userStory, bddRef, tcRef, mrTags string) error {
	_, err := s.db.Exec(`
		UPDATE api_definitions SET user_story = ?, bdd_ref = ?, tc_ref = ?, mr_tags = ?, updated_at = ?
		WHERE id = ?`,
		strings.TrimSpace(userStory), strings.TrimSpace(bddRef), strings.TrimSpace(tcRef), normalizeMRTags(mrTags),
		time.Now().UTC().Format("2006-01-02 15:04:05"), id)
	return err
}

func normalizeMRTags(raw string) string {
	parts := strings.Split(raw, ",")
	var out []string
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return strings.Join(out, ",")
}

func (s *Store) AppendMRTags(apiID int64, tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return fmt.Errorf("mr tag required")
	}
	api, err := s.GetAPI(apiID)
	if err != nil {
		return err
	}
	tags := splitMRTags(api.MRTags)
	for _, t := range tags {
		if t == tag {
			return nil
		}
	}
	tags = append(tags, tag)
	return s.UpdateAPIMeta(apiID, api.UserStory, api.BDDRef, api.TCRef, strings.Join(tags, ","))
}

func splitMRTags(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (s *Store) BulkAppendMRTag(productID int64, tag string, apiIDs []int64, paths []string) (int, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return 0, fmt.Errorf("mr tag required")
	}
	ids := map[int64]bool{}
	for _, id := range apiIDs {
		if id > 0 {
			ids[id] = true
		}
	}
	if len(paths) > 0 {
		all, err := s.ListAPIsFiltered(APIListFilter{ProductID: productID})
		if err != nil {
			return 0, err
		}
		for _, api := range all {
			for _, p := range paths {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				if strings.Contains(api.Path, p) || strings.EqualFold(api.Method+":"+api.Path, p) {
					ids[api.ID] = true
					break
				}
			}
		}
	}
	n := 0
	for id := range ids {
		if err := s.AppendMRTags(id, tag); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (s *Store) GetAPICoverage(productID int64, mrTag string) (*APICoverage, error) {
	list, err := s.ListAPIsFiltered(APIListFilter{ProductID: productID, MRTag: mrTag})
	if err != nil {
		return nil, err
	}
	c := &APICoverage{ByMR: map[string]int{}}
	for _, api := range list {
		c.Total++
		if strings.TrimSpace(api.UserStory) != "" {
			c.WithUserStory++
		} else {
			c.GapNoUS++
		}
		if strings.TrimSpace(api.BDDRef) != "" {
			c.WithBDD++
		} else {
			c.GapNoBDD++
		}
		if strings.TrimSpace(api.TCRef) != "" {
			c.WithTC++
		} else {
			c.GapNoTC++
		}
		if api.AssertionCount > 0 {
			c.WithAssertions++
		} else {
			c.GapNoAssert++
		}
		if api.ScenarioReady {
			c.ScenarioReady++
		}
		for _, t := range splitMRTags(api.MRTags) {
			c.ByMR[t]++
		}
	}
	return c, nil
}

// GetAPIByID loads full API including meta columns (extends GetAPI).
func (s *Store) getAPIRow(id int64) (*APIDefinition, error) {
	api := &APIDefinition{}
	row := s.db.QueryRow(`SELECT `+apiSelectCols+` FROM api_definitions a WHERE a.id = ?`, id)
	if err := scanAPI(row, api, nil); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return api, nil
}
