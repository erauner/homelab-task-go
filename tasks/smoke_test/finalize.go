package smoke_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/erauner/homelab-task-go/pkg/taskkit"
)

func init() {
	taskkit.Register("smoke-test-finalize", HandleFinalize)
}

// SmokeTestReport is the final report structure
type SmokeTestReport struct {
	TestName     string         `json:"test_name"`
	TaskID       string         `json:"task_id"`
	StartTime    string         `json:"start_time"`
	EndTime      string         `json:"end_time"`
	ChecksPassed bool           `json:"checks_passed"`
	GoVersion    string         `json:"go_version,omitempty"`
	Status       string         `json:"status"`
	Details      map[string]any `json:"details,omitempty"`
}

// HandleFinalize generates the smoke test report and determines final status.
func HandleFinalize(input taskkit.StepInput, deps taskkit.Deps) taskkit.StepResult {
	result := taskkit.NewStepResult()

	deps.Logger("Running smoke-test-finalize")

	// Gather vars
	testName := getStringVar(input.Vars, "test_name", "unknown")
	startTime := getStringVar(input.Vars, "start_time", "")
	checksPassed := getBoolVar(input.Vars, "checks_passed", false)
	goVersion := getStringVar(input.Vars, "go_version", "")

	// Build report
	report := SmokeTestReport{
		TestName:     testName,
		TaskID:       input.TaskID,
		StartTime:    startTime,
		EndTime:      time.Now().Format(time.RFC3339),
		ChecksPassed: checksPassed,
		GoVersion:    goVersion,
		Status:       "passed",
		Details: map[string]any{
			"workflow":    input.WorkflowName,
			"initialized": getBoolVar(input.Vars, "initialized", false),
		},
	}

	if !checksPassed {
		report.Status = "failed"
		result.AddError("Smoke test checks failed", "smoke-test")
	} else {
		result.AddInfo("All smoke test checks passed", "smoke-test")
	}

	// Write report to workdir
	if deps.Workdir != "" {
		reportPath := filepath.Join(deps.Workdir, "smoke-test-report.json")
		if data, err := json.MarshalIndent(report, "", "  "); err == nil {
			if err := os.WriteFile(reportPath, data, 0644); err != nil {
				result.AddWarning(fmt.Sprintf("Failed to write report: %v", err), "smoke-test")
			} else {
				result.AddInfo(fmt.Sprintf("Report written to: %s", reportPath), "smoke-test")
			}
		}
	}

	result.SetVar("final_status", report.Status)
	result.SetOutput("report", report)

	return result
}

func getStringVar(vars map[string]any, key, defaultVal string) string {
	if v, ok := vars[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getBoolVar(vars map[string]any, key string, defaultVal bool) bool {
	if v, ok := vars[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}
