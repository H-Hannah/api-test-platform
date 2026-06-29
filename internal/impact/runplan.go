package impact

import (
	"fmt"

	"api-test-platform/internal/runner"
	"api-test-platform/internal/store"
)

func (s *Service) ExecuteRunPlan(runnerSvc *runner.Service, req RunPlanRequest) (*RunPlanResult, error) {
	if req.EnvID <= 0 {
		return nil, fmt.Errorf("env_id required")
	}
	out := &RunPlanResult{}
	for _, apiID := range req.APIIDs {
		if req.DatasetID > 0 {
			s.runAPIOnce(runnerSvc, req.EnvID, apiID, req.DatasetID, "", out)
			continue
		}
		datasets, err := s.store.ListTestDatasets(store.TestDatasetFilter{APIID: apiID})
		if err != nil {
			s.skipAPI(apiID, out, err.Error())
			continue
		}
		if len(datasets) == 0 {
			s.skipAPI(apiID, out, "无用例，已跳过")
			continue
		}
		for _, ds := range datasets {
			s.runAPIOnce(runnerSvc, req.EnvID, apiID, ds.ID, ds.DatasetKey, out)
		}
	}
	for _, scID := range req.ScenarioIDs {
		item := RunPlanItem{Kind: "scenario", TargetID: scID}
		sc, err := s.store.GetScenario(scID)
		if err != nil {
			item.Status = "error"
			item.Error = err.Error()
			out.Runs = append(out.Runs, item)
			out.Failed++
			out.Total++
			continue
		}
		item.Name = sc.Name
		run, err := runnerSvc.RunScenario(scID, req.EnvID)
		if err != nil {
			item.Status = "error"
			item.Error = err.Error()
			out.Failed++
		} else {
			item.RunID = run.ID
			item.Status = run.Status
			if run.Status == "passed" {
				out.Passed++
			} else {
				out.Failed++
			}
		}
		out.Runs = append(out.Runs, item)
		out.Total++
	}
	return out, nil
}

func (s *Service) skipAPI(apiID int64, out *RunPlanResult, reason string) {
	item := RunPlanItem{Kind: "api", TargetID: apiID, Status: "skipped", Error: reason}
	if api, err := s.store.GetAPI(apiID); err == nil {
		item.Name = api.Name
	}
	out.Runs = append(out.Runs, item)
	out.Skipped++
	out.Total++
}

func (s *Service) runAPIOnce(runnerSvc *runner.Service, envID, apiID, datasetID int64, datasetKey string, out *RunPlanResult) {
	item := RunPlanItem{Kind: "api", TargetID: apiID, DatasetID: datasetID, DatasetKey: datasetKey}
	api, err := s.store.GetAPI(apiID)
	if err != nil {
		item.Status = "error"
		item.Error = err.Error()
		out.Runs = append(out.Runs, item)
		out.Failed++
		out.Total++
		return
	}
	item.Name = api.Name
	if datasetKey != "" {
		item.Name = api.Name + " · " + datasetKey
	}
	run, err := runnerSvc.RunAPI(apiID, envID, datasetID)
	if err != nil {
		item.Status = "error"
		item.Error = err.Error()
		out.Failed++
	} else {
		item.RunID = run.ID
		item.Status = run.Status
		if run.Status == "passed" {
			out.Passed++
		} else {
			out.Failed++
		}
	}
	out.Runs = append(out.Runs, item)
	out.Total++
}
