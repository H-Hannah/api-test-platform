package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"api-test-platform/internal/ai"
	"api-test-platform/internal/config"
	"api-test-platform/internal/docsrepo"
	"api-test-platform/internal/impact"
	"api-test-platform/internal/runner"
	"api-test-platform/internal/store"
)

type Server struct {
	cfg       config.Config
	store     *store.Store
	ingest    *ai.IngestService
	testData  *ai.TestDataService
	impact    *impact.Service
	mr        *ai.MRService
	runner    *runner.Service
}

func NewServer(cfg config.Config, st *store.Store) *Server {
	aiClient := ai.NewClient(cfg.AIAPIKey, cfg.AIBaseURL, cfg.AIModel, cfg.AIVisionModel)
	ingest := ai.NewIngestService(aiClient, st)
	return &Server{
		cfg:       cfg,
		store:     st,
		ingest:    ingest,
		testData:  ai.NewTestDataService(aiClient, st),
		impact:    impact.NewService(st, cfg.DocsRepoRoot, aiClient),
		mr:        ai.NewMRService(aiClient, ingest, st, cfg.DocsRepoRoot),
		runner:    runner.New(st),
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(AuthMiddleware(s.cfg.APIToken))

	r.Get("/health", s.health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/products", s.listProducts)

		r.Get("/environments", s.listAllEnvironments)
		r.Post("/environments", s.createGlobalEnvironment)
		r.Get("/folders/tree", s.folderTreeGlobal)
		r.Get("/apis", s.listAPIsGlobal)
		r.Get("/scenarios", s.listScenariosGlobal)
		r.Get("/testdata/datasets", s.listTestDatasetsGlobal)
		r.Post("/testdata/import", s.importTestDataGlobal)

		r.Get("/products/{productId}/folders/tree", s.folderTree)
		r.Get("/products/{productId}/folders", s.listFolders)
		r.Post("/products/{productId}/folders", s.createFolder)

		r.Get("/products/{productId}/environments", s.listEnvironments)
		r.Post("/products/{productId}/environments", s.createEnvironment)

		r.Get("/environments/{id}", s.getEnvironment)
		r.Put("/environments/{id}", s.updateEnvironment)
		r.Delete("/environments/{id}", s.deleteEnvironment)

		r.Get("/products/{productId}/apis", s.listAPIs)
		r.Get("/products/{productId}/coverage", s.getAPICoverage)
		r.Post("/products/{productId}/apis/bulk-mr-tag", s.bulkMRTag)
		r.Get("/apis/{id}", s.getAPI)
		r.Patch("/apis/{id}/meta", s.patchAPIMeta)
		r.Delete("/apis/{id}", s.deleteAPI)
		r.Post("/apis/{id}/generate-cases", s.generateAPICases)
		r.Post("/apis/{id}/run", s.runAPI)
		r.Get("/apis/{id}/runs", s.listAPIRuns)

		r.Get("/products/{productId}/scenarios", s.listScenarios)
		r.Get("/scenarios/{id}", s.getScenario)
		r.Post("/scenarios/{id}/run", s.runScenario)

		r.Get("/runs", s.listRuns)
		r.Get("/runs/{id}", s.getRun)

		r.Post("/ai/ingest", s.aiIngest)
		r.Post("/ai/testdata/generate", s.aiGenerateTestData)
		r.Post("/products/{productId}/testdata/import", s.importTestData)
		r.Get("/products/{productId}/testdata/datasets", s.listTestDatasets)
		r.Get("/testdata/datasets/{id}", s.getTestDataset)
		r.Patch("/testdata/datasets/{id}", s.patchTestDataset)
		r.Delete("/testdata/datasets/{id}", s.deleteTestDataset)
		r.Post("/environments/{id}/import-var-keys", s.importEnvVarKeys)
		r.Post("/impact/analyze", s.impactAnalyze)
		r.Post("/impact/post-mr-comment", s.impactPostMRComment)
		r.Post("/impact/preview-mr-comment", s.impactPreviewMRComment)
		r.Post("/impact/run-plan", s.impactRunPlan)
		r.Post("/ai/mr/verify-tc", s.aiVerifyMRWithTC)
		r.Post("/ai/mr/ingest-apis", s.aiIngestMRAPIs)
		r.Get("/docs/testcases/branches", s.listDocsBranches)
		r.Get("/docs/testcases/catalog", s.listTestDocsCatalog)
		r.Get("/docs/testcases", s.loadTestCases)
		r.Get("/docs/requirement-package", s.loadRequirementPackage)
	})

	if _, err := os.Stat(WebDistDir()); err == nil {
		r.Handle("/*", SPAHandler(WebDistDir()))
	}

	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) aiIngest(w http.ResponseWriter, r *http.Request) {
	var req ai.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := s.ingest.Ingest(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, resp)
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListProducts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.Product{}
	}
	writeJSON(w, list)
}

func (s *Server) listAllEnvironments(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListAllEnvironments()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) createGlobalEnvironment(w http.ResponseWriter, r *http.Request) {
	s.createEnvironmentBody(w, r)
}

func (s *Server) folderTreeGlobal(w http.ResponseWriter, r *http.Request) {
	tree, err := s.store.BuildFolderTree(store.AllProducts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, tree)
}

func (s *Server) listAPIsGlobal(w http.ResponseWriter, r *http.Request) {
	s.listAPIsWithProduct(w, r, store.AllProducts)
}

func (s *Server) listScenariosGlobal(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListScenarios(store.AllProducts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) listTestDatasetsGlobal(w http.ResponseWriter, r *http.Request) {
	s.listTestDatasetsWithProduct(w, r, store.AllProducts)
}

func (s *Server) importTestDataGlobal(w http.ResponseWriter, r *http.Request) {
	s.importTestDataWithProduct(w, r, store.ResolveProductID(0))
}

func (s *Server) folderTree(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	tree, err := s.store.BuildFolderTree(pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, tree)
}

func (s *Server) listFolders(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	list, err := s.store.ListFolders(pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) createFolder(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	var body struct {
		ParentID int64  `json:"parent_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	f, err := s.store.CreateFolder(pid, body.ParentID, body.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, f)
}

func (s *Server) listEnvironments(w http.ResponseWriter, r *http.Request) {
	s.listAllEnvironments(w, r)
}

func (s *Server) createEnvironment(w http.ResponseWriter, r *http.Request) {
	s.createEnvironmentBody(w, r)
}

func (s *Server) createEnvironmentBody(w http.ResponseWriter, r *http.Request) {
	var e store.Environment
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(e.Name) == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}
	if strings.TrimSpace(e.BaseURL) == "" {
		writeError(w, http.StatusBadRequest, "base_url required")
		return
	}
	vars, err := store.NormalizeEnvVariablesJSON(e.BaseURL, e.Variables)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	e.Variables = vars
	id, err := s.store.CreateEnvironment(&e)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if e.IsDefault {
		_ = s.store.ClearDefaultExcept(id)
	}
	got, _ := s.store.GetEnvironment(id)
	if got != nil {
		writeJSON(w, got)
		return
	}
	e.ID = id
	writeJSON(w, e)
}

func (s *Server) getEnvironment(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	e, err := s.store.GetEnvironment(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, e)
}

func (s *Server) updateEnvironment(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	var body store.Environment
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	existing, err := s.store.GetEnvironment(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	body.ID = id
	if strings.TrimSpace(body.Name) == "" {
		body.Name = existing.Name
	}
	if strings.TrimSpace(body.BaseURL) == "" {
		writeError(w, http.StatusBadRequest, "base_url required")
		return
	}
	if err := s.store.UpdateEnvironment(&body); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "environment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	e, _ := s.store.GetEnvironment(id)
	writeJSON(w, e)
}

func (s *Server) deleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	if _, err := s.store.GetEnvironment(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := s.store.DeleteEnvironment(id); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "environment not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) listAPIs(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	s.listAPIsWithProduct(w, r, pid)
}

func (s *Server) listAPIsWithProduct(w http.ResponseWriter, r *http.Request, pid int64) {
	fid, _ := strconv.ParseInt(r.URL.Query().Get("folder_id"), 10, 64)
	f := store.APIListFilter{
		ProductID: pid,
		FolderID:  fid,
		UserStory: r.URL.Query().Get("user_story"),
		MRTag:     r.URL.Query().Get("mr_tag"),
		Gap:       r.URL.Query().Get("gap"),
	}
	list, err := s.store.ListAPIsFiltered(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) getAPICoverage(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	c, err := s.store.GetAPICoverage(pid, r.URL.Query().Get("mr_tag"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, c)
}

func (s *Server) patchAPIMeta(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	var body struct {
		UserStory string `json:"user_story"`
		BDDRef    string `json:"bdd_ref"`
		TCRef     string `json:"tc_ref"`
		MRTags    string `json:"mr_tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.store.UpdateAPIMeta(id, body.UserStory, body.BDDRef, body.TCRef, body.MRTags); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	api, err := s.store.GetAPI(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, api)
}

func (s *Server) bulkMRTag(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	var body struct {
		MRTag  string   `json:"mr_tag"`
		APIIDs []int64  `json:"api_ids"`
		Paths  []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	n, err := s.store.BulkAppendMRTag(pid, body.MRTag, body.APIIDs, body.Paths)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true, "tagged": n, "mr_tag": body.MRTag})
}

func (s *Server) getAPI(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	api, err := s.store.GetAPI(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, api)
}

func (s *Server) generateAPICases(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	var body struct {
		Hint string `json:"hint"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	resp, err := s.ingest.GenerateAPICases(id, body.Hint)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, resp)
}

func (s *Server) deleteAPI(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	if err := s.store.DeleteAPI(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) runAPI(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	var body struct {
		EnvID     int64 `json:"env_id"`
		DatasetID int64 `json:"dataset_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.EnvID == 0 {
		writeError(w, http.StatusBadRequest, "env_id required")
		return
	}
	run, err := s.runner.RunAPI(id, body.EnvID, body.DatasetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, run)
}

func (s *Server) listScenarios(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	list, err := s.store.ListScenarios(pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) getScenario(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	sc, err := s.store.GetScenario(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, sc)
}

func (s *Server) runScenario(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	var body struct {
		EnvID int64 `json:"env_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.EnvID == 0 {
		writeError(w, http.StatusBadRequest, "env_id required")
		return
	}
	run, err := s.runner.RunScenario(id, body.EnvID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, run)
}

func (s *Server) listAPIRuns(w http.ResponseWriter, r *http.Request) {
	apiID := paramInt64(r, "id")
	if apiID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid api id")
		return
	}
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	list, err := s.store.ListRunsByAPI(apiID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListRuns(50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, list)
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	run, err := s.store.GetRun(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, run)
}

func (s *Server) listDocsBranches(w http.ResponseWriter, r *http.Request) {
	br, err := docsrepo.ListDocsBranches()
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, br)
}

func (s *Server) listTestDocsCatalog(w http.ResponseWriter, r *http.Request) {
	version := strings.TrimSpace(r.URL.Query().Get("version"))
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	cat, err := docsrepo.ListTestDocsCatalog(version, ref)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, cat)
}

func (s *Server) loadTestCases(w http.ResponseWriter, r *http.Request) {
	version := strings.TrimSpace(r.URL.Query().Get("version"))
	requirementID := strings.TrimSpace(r.URL.Query().Get("requirement_id"))
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if version != "" && requirementID != "" {
		tc, err := docsrepo.LoadTestCasesBySelector(version, requirementID, ref)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSON(w, tc)
		return
	}
	gitlabURL := strings.TrimSpace(r.URL.Query().Get("gitlab_url"))
	if gitlabURL == "" {
		writeError(w, http.StatusBadRequest, "version+requirement_id 或 gitlab_url required")
		return
	}
	tc, err := docsrepo.LoadTestCasesFromGitLabURL(gitlabURL)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, tc)
}

func (s *Server) loadRequirementPackage(w http.ResponseWriter, r *http.Request) {
	version := strings.TrimSpace(r.URL.Query().Get("version"))
	requirementID := strings.TrimSpace(r.URL.Query().Get("requirement_id"))
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if version == "" || requirementID == "" {
		writeError(w, http.StatusBadRequest, "version and requirement_id required")
		return
	}
	pkg, err := docsrepo.LoadRequirementPackage(version, requirementID, ref)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, pkg)
}

func (s *Server) aiVerifyMRWithTC(w http.ResponseWriter, r *http.Request) {
	var req ai.MRVerifyTCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.mr.VerifyMRWithTC(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) aiIngestMRAPIs(w http.ResponseWriter, r *http.Request) {
	var req ai.MRIngestAPIsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.mr.IngestAPIsFromMR(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) aiGenerateTestData(w http.ResponseWriter, r *http.Request) {
	var req ai.TestDataGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.testData.Generate(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) importTestData(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	s.importTestDataWithProduct(w, r, pid)
}

func (s *Server) importTestDataWithProduct(w http.ResponseWriter, r *http.Request, pid int64) {
	var req ai.TestDataImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.ProductID = store.ResolveProductID(req.ProductID)
	if pid > 0 {
		req.ProductID = pid
	}
	res, err := s.testData.Import(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) listTestDatasets(w http.ResponseWriter, r *http.Request) {
	pid := paramInt64(r, "productId")
	s.listTestDatasetsWithProduct(w, r, pid)
}

func (s *Server) listTestDatasetsWithProduct(w http.ResponseWriter, r *http.Request, pid int64) {
	q := r.URL.Query()
	apiID, _ := strconv.ParseInt(q.Get("api_id"), 10, 64)
	list, err := s.store.ListTestDatasets(store.TestDatasetFilter{
		ProductID:     pid,
		Version:       q.Get("version"),
		RequirementID: q.Get("requirement_id"),
		APIID:         apiID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []store.TestDataset{}
	}
	if apiID > 0 {
		for i := range list {
			s.store.EnrichDatasetStale(apiID, &list[i])
		}
	}
	writeJSON(w, list)
}

func (s *Server) getTestDataset(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	ds, err := s.store.GetTestDataset(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, ds)
}

func (s *Server) patchTestDataset(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	var body struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Variables       string `json:"variables"`
		HeadersOverride string `json:"headers_override"`
		BodyOverride    string `json:"body_override"`
		Assertions      string `json:"assertions"`
		Tags            string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	existing, err := s.store.GetTestDataset(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if v := strings.TrimSpace(body.Name); v != "" {
		existing.Name = v
	}
	existing.Description = body.Description
	if strings.TrimSpace(body.Variables) != "" {
		existing.Variables = body.Variables
	} else if body.Variables != "" {
		existing.Variables = "{}"
	}
	if body.HeadersOverride != "" {
		existing.HeadersOverride = body.HeadersOverride
	}
	existing.BodyOverride = body.BodyOverride
	if strings.TrimSpace(body.Assertions) != "" {
		existing.Assertions = body.Assertions
	}
	if strings.TrimSpace(body.Tags) != "" {
		existing.Tags = body.Tags
	}
	if apiID := store.ParseAPIRequirementID(existing.RequirementID); apiID > 0 {
		if api, err := s.store.GetAPI(apiID); err == nil {
			existing.ApiFingerprint = store.APIDefinitionFingerprint(api)
		}
	}
	if err := s.store.UpdateTestDataset(existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := s.store.GetTestDataset(id)
	writeJSON(w, updated)
}

func (s *Server) deleteTestDataset(w http.ResponseWriter, r *http.Request) {
	id := paramInt64(r, "id")
	if err := s.store.DeleteTestDataset(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) importEnvVarKeys(w http.ResponseWriter, r *http.Request) {
	envID := paramInt64(r, "id")
	var req ai.ImportEnvKeysRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.testData.ImportEnvKeys(envID, req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) impactAnalyze(w http.ResponseWriter, r *http.Request) {
	var req impact.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.impact.Analyze(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) impactPostMRComment(w http.ResponseWriter, r *http.Request) {
	var req impact.PostMRCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.impact.PostMRComment(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) impactPreviewMRComment(w http.ResponseWriter, r *http.Request) {
	var req impact.PostMRCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.impact.PreviewMRComment(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func (s *Server) impactRunPlan(w http.ResponseWriter, r *http.Request) {
	var req impact.RunPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.impact.ExecuteRunPlan(s.runner, req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, res)
}

func paramInt64(r *http.Request, key string) int64 {
	v, _ := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	return v
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
