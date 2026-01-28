// Package smoke_test implements a basic smoke test workflow for verifying taskkit functionality.
package smoke_test

import (
	"fmt"
	"time"

	"github.com/erauner/homelab-task-go/pkg/taskkit"
)

func init() {
	taskkit.Register("smoke-test-init", HandleInit)
}

// HandleInit validates input parameters and initializes workflow variables.
func HandleInit(input taskkit.StepInput, deps taskkit.Deps) taskkit.StepResult {
	result := taskkit.NewStepResult()

	deps.Logger("Starting smoke-test-init")

	// Validate required params
	testName := input.GetParamString("test_name")
	if testName == "" {
		testName = "default-smoke-test"
	}

	// Initialize workflow variables
	result.SetVar("test_name", testName)
	result.SetVar("start_time", time.Now().Format(time.RFC3339))
	result.SetVar("initialized", true)

	result.AddInfo(fmt.Sprintf("Smoke test initialized: %s", testName), "smoke-test")
	result.AddInfo(fmt.Sprintf("Task ID: %s", input.TaskID), "smoke-test")
	result.AddInfo(fmt.Sprintf("Workflow: %s", input.WorkflowName), "smoke-test")

	return result
}
