package taskkit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// LocalRunnerConfig holds configuration for the runner
type LocalRunnerConfig struct {
	WorkflowPath string
	ParamsPath   string
	Workdir      string
	TaskID       string
	Verbose      bool
}

// LocalRunner executes workflows locally
type LocalRunner struct {
	config   LocalRunnerConfig
	workflow *WorkflowDefinition
	params   map[string]any
	vars     map[string]any
	deps     Deps
}

// NewLocalRunner creates a new runner instance
func NewLocalRunner(config LocalRunnerConfig) (*LocalRunner, error) {
	// Load workflow
	wf, err := LoadWorkflow(config.WorkflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}

	// Load params
	params := make(map[string]any)
	if config.ParamsPath != "" {
		data, err := os.ReadFile(config.ParamsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read params file: %w", err)
		}
		if err := json.Unmarshal(data, &params); err != nil {
			return nil, fmt.Errorf("failed to parse params JSON: %w", err)
		}
	}

	// Ensure workdir exists
	if config.Workdir == "" {
		config.Workdir = "."
	}
	if err := os.MkdirAll(config.Workdir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workdir: %w", err)
	}

	// Load existing vars if present
	vars := make(map[string]any)
	varsPath := filepath.Join(config.Workdir, "vars.yaml")
	if data, err := os.ReadFile(varsPath); err == nil {
		yaml.Unmarshal(data, &vars)
	}

	logger := func(format string, args ...any) {
		if config.Verbose {
			fmt.Printf("[DEBUG] "+format+"\n", args...)
		}
	}

	return &LocalRunner{
		config:   config,
		workflow: wf,
		params:   params,
		vars:     vars,
		deps: Deps{
			Workdir: config.Workdir,
			Logger:  logger,
		},
	}, nil
}

// Run executes the workflow and returns the final result
func (r *LocalRunner) Run() ExecutionResult {
	startTime := time.Now()

	result := ExecutionResult{
		TaskID:       r.config.TaskID,
		WorkflowName: r.workflow.Name,
		StartTime:    startTime,
		Steps:        make([]StepExec, 0),
	}

	// Get execution order
	steps, err := r.workflow.GetExecutionOrder()
	if err != nil {
		result.Result = "Error"
		result.ErrorMessage = fmt.Sprintf("Failed to determine execution order: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime).String()
		r.saveResult(result)
		return result
	}

	fmt.Printf("=== Executing workflow: %s ===\n", r.workflow.Name)
	fmt.Printf("Steps: %d\n", len(steps))

	// Execute each step
	workflowFailed := false
	for _, step := range steps {
		stepExec := r.executeStep(step)
		result.Steps = append(result.Steps, stepExec)

		if stepExec.Status == "Failed" {
			workflowFailed = true
			// Check if this is the finalize step - if so, continue to record the result
			if step.Template != TemplateFinalize {
				// For non-finalize steps, check if there's a finalize step to run
				hasFinalize := false
				for _, s := range steps {
					if s.Template == TemplateFinalize {
						hasFinalize = true
						break
					}
				}
				if !hasFinalize {
					break
				}
			}
		}
	}

	// Determine final result
	if workflowFailed {
		result.Result = "Failed"
	} else {
		result.Result = "Succeeded"
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime).String()
	result.FinalVars = r.vars

	// Save results
	r.saveResult(result)
	r.saveVars()

	fmt.Printf("\n=== Workflow %s: %s ===\n", r.workflow.Name, result.Result)
	return result
}

func (r *LocalRunner) executeStep(step WorkflowStep) StepExec {
	stepStart := time.Now()
	handlerName := r.workflow.GetHandlerName(step)

	exec := StepExec{
		Name:    step.Name,
		Handler: handlerName,
	}

	fmt.Printf("\n--- Step: %s (handler: %s) ---\n", step.Name, handlerName)

	// Get handler
	handler, ok := Get(handlerName)
	if !ok {
		exec.Status = "Failed"
		exec.Error = fmt.Sprintf("handler not found: %s", handlerName)
		exec.Duration = time.Since(stepStart).String()
		fmt.Printf("ERROR: %s\n", exec.Error)
		return exec
	}

	// Build input
	input := StepInput{
		StepName:     step.Name,
		TaskID:       r.config.TaskID,
		WorkflowName: r.workflow.Name,
		Attempt:      1,
		TotalRetries: r.workflow.GetRetries(step),
		Params:       r.mergeParams(step.Params),
		Vars:         r.vars,
	}

	// Execute with retries
	maxAttempts := r.workflow.GetRetries(step) + 1
	var stepResult StepResult

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		input.Attempt = attempt

		if attempt > 1 {
			fmt.Printf("  Retry attempt %d/%d\n", attempt, maxAttempts)
		}

		stepResult = handler(input, r.deps)

		// Check for skip
		if skip, ok := stepResult.FlowControl["skip"].(bool); ok && skip {
			exec.Status = "Skipped"
			reason := ""
			if r, ok := stepResult.FlowControl["skip_reason"].(string); ok {
				reason = r
			}
			exec.Error = reason
			break
		}

		// Check for errors
		if !stepResult.HasErrors() {
			exec.Status = "Succeeded"
			break
		}

		// Last attempt failed
		if attempt == maxAttempts {
			exec.Status = "Failed"
		}
	}

	// Record results
	exec.Messages = stepResult.Messages
	exec.Output = stepResult.Output
	exec.Duration = time.Since(stepStart).String()

	// Apply context updates to vars
	for k, v := range stepResult.ContextUpdates {
		r.vars[k] = v
	}

	// Print messages
	for _, msg := range stepResult.Messages {
		fmt.Printf("  [%s] %s\n", msg.Severity, msg.Text)
	}

	fmt.Printf("  Status: %s (duration: %s)\n", exec.Status, exec.Duration)
	return exec
}

func (r *LocalRunner) mergeParams(stepParams map[string]any) map[string]any {
	merged := make(map[string]any)
	for k, v := range r.params {
		merged[k] = v
	}
	for k, v := range stepParams {
		merged[k] = v
	}
	return merged
}

func (r *LocalRunner) saveResult(result ExecutionResult) {
	path := filepath.Join(r.config.Workdir, "execution-result.json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Warning: failed to marshal result: %v\n", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("Warning: failed to write result: %v\n", err)
	}
}

func (r *LocalRunner) saveVars() {
	path := filepath.Join(r.config.Workdir, "vars.yaml")
	data, err := yaml.Marshal(r.vars)
	if err != nil {
		fmt.Printf("Warning: failed to marshal vars: %v\n", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("Warning: failed to write vars: %v\n", err)
	}
}
