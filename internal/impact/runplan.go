package impact

import (
	"fmt"

	"api-test-platform/internal/runner"
)

func (s *Service) ExecuteRunPlan(runnerSvc *runner.Service, req RunPlanRequest) (*RunPlanResult, error) {
	if req.EnvID <= 0 {
		return nil, fmt.Errorf("env_id required")
	}
	out := &RunPlanResult{}
	for _, apiID := range req.APIIDs {
		item := RunPlanItem{Kind: "api", TargetID: apiID}
		api, err := s.store.GetAPI(apiID)
		if err != nil {
			item.Status = "error"
			item.Error = err.Error()
			out.Runs = append(out.Runs, item)
			out.Failed++
			continue
		}
		item.Name = api.Name
		run, err := runnerSvc.RunAPI(apiID, req.EnvID, req.DatasetID)
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
	for _, scID := range req.ScenarioIDs {
		item := RunPlanItem{Kind: "scenario", TargetID: scID}
		sc, err := s.store.GetScenario(scID)
		if err != nil {
			item.Status = "error"
			item.Error = err.Error()
			out.Runs = append(out.Runs, item)
			out.Failed++
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
