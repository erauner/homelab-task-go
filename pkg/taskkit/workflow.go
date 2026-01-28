package taskkit

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// StepTemplate defines standard step types
type StepTemplate string

const (
	TemplateInit     StepTemplate = "init"
	TemplateAction   StepTemplate = "action"
	TemplateFinalize StepTemplate = "finalize"
)

// WorkflowStep defines a single step in a workflow
type WorkflowStep struct {
	Name     string         `yaml:"name"`
	Depends  []string       `yaml:"depends,omitempty"`
	Template StepTemplate   `yaml:"template,omitempty"`
	Params   map[string]any `yaml:"params,omitempty"`
	Retries  int            `yaml:"retries,omitempty"`
}

// WorkflowDefinition is the parsed workflow YAML
type WorkflowDefinition struct {
	Name           string         `yaml:"name"`
	Description    string         `yaml:"description,omitempty"`
	Platform       string         `yaml:"platform"`
	HandlerPrefix  string         `yaml:"handler_prefix,omitempty"`
	Steps          []WorkflowStep `yaml:"steps"`
	DefaultRetries int            `yaml:"default_retries,omitempty"`
	TimeoutSeconds int            `yaml:"timeout_seconds,omitempty"`
}

// LoadWorkflow reads and parses a workflow YAML file
func LoadWorkflow(path string) (*WorkflowDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var wf WorkflowDefinition
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// Validate
	if wf.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if len(wf.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	return &wf, nil
}

// GetHandlerName returns the full handler name for a step
func (w *WorkflowDefinition) GetHandlerName(step WorkflowStep) string {
	// If handler_prefix is set, use prefix-stepname
	if w.HandlerPrefix != "" {
		return fmt.Sprintf("%s-%s", w.HandlerPrefix, step.Name)
	}
	// Otherwise use platform-stepname
	return fmt.Sprintf("%s-%s", w.Platform, step.Name)
}

// GetExecutionOrder returns steps in topologically sorted order
// Uses Kahn's algorithm for dependency resolution
func (w *WorkflowDefinition) GetExecutionOrder() ([]WorkflowStep, error) {
	// Build adjacency list and in-degree map
	stepMap := make(map[string]WorkflowStep)
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, step := range w.Steps {
		stepMap[step.Name] = step
		inDegree[step.Name] = len(step.Depends)

		for _, dep := range step.Depends {
			dependents[dep] = append(dependents[dep], step.Name)
		}
	}

	// Validate all dependencies exist
	for _, step := range w.Steps {
		for _, dep := range step.Depends {
			if _, exists := stepMap[dep]; !exists {
				return nil, fmt.Errorf("step %q depends on unknown step %q", step.Name, dep)
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var order []WorkflowStep
	for len(queue) > 0 {
		// Pop from queue
		name := queue[0]
		queue = queue[1:]

		order = append(order, stepMap[name])

		// Reduce in-degree for dependents
		for _, dependent := range dependents[name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(order) != len(w.Steps) {
		return nil, fmt.Errorf("workflow contains a dependency cycle")
	}

	return order, nil
}

// GetRetries returns the number of retries for a step
func (w *WorkflowDefinition) GetRetries(step WorkflowStep) int {
	if step.Retries > 0 {
		return step.Retries
	}
	if w.DefaultRetries > 0 {
		return w.DefaultRetries
	}
	return 0
}
