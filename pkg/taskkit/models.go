// Package taskkit provides the core types and runtime for homelab task execution.
package taskkit

import (
	"encoding/json"
	"time"
)

// Severity levels for step messages
type Severity string

const (
	SeverityInfo    Severity = "INFO"
	SeverityWarning Severity = "WARNING"
	SeverityError   Severity = "ERROR"
	SeverityDebug   Severity = "DEBUG"
)

// Message represents a log message from a step execution
type Message struct {
	Severity  Severity  `json:"severity"`
	Text      string    `json:"text"`
	System    string    `json:"system,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// StepInput contains all context passed to a step handler
type StepInput struct {
	StepName       string         `json:"step_name"`
	TaskID         string         `json:"task_id"`
	WorkflowName   string         `json:"workflow_name"`
	Attempt        int            `json:"attempt"`
	TotalRetries   int            `json:"total_retries"`
	Params         map[string]any `json:"params"`
	Vars           map[string]any `json:"vars"`
	WorkflowResult string         `json:"workflow_result,omitempty"`
}

// GetParam retrieves a parameter by key, returning the zero value if not found
func (s *StepInput) GetParam(key string) any {
	if s.Params == nil {
		return nil
	}
	return s.Params[key]
}

// GetParamString retrieves a string parameter, returning empty string if not found
func (s *StepInput) GetParamString(key string) string {
	v := s.GetParam(key)
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

// GetVar retrieves a workflow variable by key
func (s *StepInput) GetVar(key string) any {
	if s.Vars == nil {
		return nil
	}
	return s.Vars[key]
}

// StepResult is the return type from step handlers
type StepResult struct {
	Messages       []Message      `json:"messages"`
	ContextUpdates map[string]any `json:"context_updates,omitempty"`
	Output         map[string]any `json:"output,omitempty"`
	FlowControl    map[string]any `json:"flow_control,omitempty"`
}

// NewStepResult creates a new StepResult with initialized fields
func NewStepResult() StepResult {
	return StepResult{
		Messages:       make([]Message, 0),
		ContextUpdates: make(map[string]any),
		Output:         make(map[string]any),
		FlowControl:    make(map[string]any),
	}
}

// AddMessage adds a message to the result
func (r *StepResult) AddMessage(severity Severity, text, system string) {
	r.Messages = append(r.Messages, Message{
		Severity:  severity,
		Text:      text,
		System:    system,
		Timestamp: time.Now(),
	})
}

// AddInfo adds an info message
func (r *StepResult) AddInfo(text, system string) {
	r.AddMessage(SeverityInfo, text, system)
}

// AddWarning adds a warning message
func (r *StepResult) AddWarning(text, system string) {
	r.AddMessage(SeverityWarning, text, system)
}

// AddError adds an error message
func (r *StepResult) AddError(text, system string) {
	r.AddMessage(SeverityError, text, system)
}

// AddDebug adds a debug message
func (r *StepResult) AddDebug(text, system string) {
	r.AddMessage(SeverityDebug, text, system)
}

// HasErrors returns true if the result contains any error messages
func (r *StepResult) HasErrors() bool {
	for _, m := range r.Messages {
		if m.Severity == SeverityError {
			return true
		}
	}
	return false
}

// SetVar sets a workflow variable (will be persisted to vars.yaml)
func (r *StepResult) SetVar(key string, value any) {
	r.ContextUpdates[key] = value
}

// SetOutput sets an output value (available to dependent steps)
func (r *StepResult) SetOutput(key string, value any) {
	r.Output[key] = value
}

// Skip marks the step as skipped with a reason
func (r *StepResult) Skip(reason string) {
	r.FlowControl["skip"] = true
	r.FlowControl["skip_reason"] = reason
}

// ExecutionResult is the final result of a workflow execution
type ExecutionResult struct {
	Result       string         `json:"result"` // Succeeded, Failed, Error
	TaskID       string         `json:"task_id"`
	WorkflowName string         `json:"workflow_name"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	Duration     string         `json:"duration"`
	Steps        []StepExec     `json:"steps"`
	FinalVars    map[string]any `json:"final_vars,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

// StepExec records the execution of a single step
type StepExec struct {
	Name      string        `json:"name"`
	Handler   string        `json:"handler"`
	Status    string        `json:"status"` // Succeeded, Failed, Skipped
	Duration  string        `json:"duration"`
	Messages  []Message     `json:"messages,omitempty"`
	Output    map[string]any `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// Deps provides external dependencies to step handlers
type Deps struct {
	Workdir string
	Logger  func(format string, args ...any)
}

// ToJSON serializes any value to JSON string
func ToJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
