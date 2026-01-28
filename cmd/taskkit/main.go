// Command taskkit runs homelab task workflows.
//
// Usage:
//
//	taskkit workflow run --workflow <path> [options]
//	taskkit list-handlers
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erauner/homelab-task-go/pkg/taskkit"
	// Import task packages to register handlers via init()
	_ "github.com/erauner/homelab-task-go/tasks/smoke_test"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "workflow":
		if len(os.Args) < 3 || os.Args[2] != "run" {
			fmt.Println("Usage: taskkit workflow run --workflow <path> [options]")
			os.Exit(1)
		}
		runWorkflow(os.Args[3:])

	case "list-handlers":
		listHandlers()

	case "version":
		fmt.Println("taskkit v0.1.0")

	case "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`taskkit - Homelab task workflow runner

Commands:
  workflow run    Execute a workflow
  list-handlers   List all registered step handlers
  version         Show version

Workflow Options:
  --workflow, -w  Path to workflow YAML file (required)
  --params, -p    Path to params.json file
  --workdir       Working directory for outputs
  --task-id       Task ID for tracking
  --verbose, -v   Enable verbose logging

Example:
  taskkit workflow run --workflow workflows/smoke_test.yaml --workdir /tmp/run`)
}

func runWorkflow(args []string) {
	fs := flag.NewFlagSet("workflow run", flag.ExitOnError)
	workflowPath := fs.String("workflow", "", "Path to workflow YAML file")
	fs.StringVar(workflowPath, "w", "", "Path to workflow YAML file (shorthand)")
	paramsPath := fs.String("params", "", "Path to params.json file")
	fs.StringVar(paramsPath, "p", "", "Path to params.json file (shorthand)")
	workdir := fs.String("workdir", "", "Working directory for outputs")
	taskID := fs.String("task-id", "", "Task ID for tracking")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	fs.BoolVar(verbose, "v", false, "Enable verbose logging (shorthand)")

	if err := fs.Parse(args); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if *workflowPath == "" {
		fmt.Println("Error: --workflow is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	config := taskkit.LocalRunnerConfig{
		WorkflowPath: *workflowPath,
		ParamsPath:   *paramsPath,
		Workdir:      *workdir,
		TaskID:       *taskID,
		Verbose:      *verbose,
	}

	runner, err := taskkit.NewLocalRunner(config)
	if err != nil {
		fmt.Printf("Error initializing runner: %v\n", err)
		os.Exit(1)
	}

	result := runner.Run()

	// Exit with appropriate code
	switch result.Result {
	case "Succeeded":
		os.Exit(0)
	case "Failed":
		os.Exit(1)
	default:
		os.Exit(2)
	}
}

func listHandlers() {
	handlers := taskkit.ListHandlers()
	fmt.Printf("Registered step handlers (%d):\n", len(handlers))
	for _, name := range handlers {
		fmt.Printf("  - %s\n", name)
	}
}
