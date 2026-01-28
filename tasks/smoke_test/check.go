package smoke_test

import (
	"fmt"
	"os"
	"runtime"

	"github.com/erauner/homelab-task-go/pkg/taskkit"
)

func init() {
	taskkit.Register("smoke-test-check", HandleCheck)
}

// HandleCheck performs basic system checks and records results.
func HandleCheck(input taskkit.StepInput, deps taskkit.Deps) taskkit.StepResult {
	result := taskkit.NewStepResult()

	deps.Logger("Running smoke-test-check")

	// Get test name from vars
	testName := "unknown"
	if v := input.GetVar("test_name"); v != nil {
		if s, ok := v.(string); ok {
			testName = s
		}
	}

	result.AddInfo(fmt.Sprintf("Running checks for: %s", testName), "smoke-test")

	// Check 1: Go runtime
	result.AddInfo(fmt.Sprintf("Go version: %s", runtime.Version()), "smoke-test")
	result.AddInfo(fmt.Sprintf("GOOS: %s, GOARCH: %s", runtime.GOOS, runtime.GOARCH), "smoke-test")

	// Check 2: Working directory
	if deps.Workdir != "" {
		if info, err := os.Stat(deps.Workdir); err == nil && info.IsDir() {
			result.AddInfo(fmt.Sprintf("Workdir exists: %s", deps.Workdir), "smoke-test")
		} else {
			result.AddWarning(fmt.Sprintf("Workdir issue: %s", deps.Workdir), "smoke-test")
		}
	}

	// Check 3: Environment
	if hostname, err := os.Hostname(); err == nil {
		result.AddInfo(fmt.Sprintf("Hostname: %s", hostname), "smoke-test")
	}

	// Record check results
	result.SetVar("checks_passed", true)
	result.SetVar("go_version", runtime.Version())

	result.SetOutput("runtime", map[string]any{
		"go_version": runtime.Version(),
		"goos":       runtime.GOOS,
		"goarch":     runtime.GOARCH,
	})

	return result
}
