// Package smoke_test provides step handlers for the smoke-test workflow.
//
// This package demonstrates the taskkit step registration pattern.
// Each step handler is registered in its init() function using taskkit.Register().
//
// Handlers:
//   - smoke-test-init: Validates parameters and initializes workflow variables
//   - smoke-test-check: Performs basic system checks
//   - smoke-test-finalize: Generates final report and status
//
// Usage:
//
//	Import this package in cmd/taskkit/main.go to register handlers:
//
//	    import _ "github.com/erauner/homelab-task-go/tasks/smoke_test"
package smoke_test
