package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AgentStatus represents the current state of an agent
type AgentStatus struct {
	Rule      string    `json:"rule"`
	Status    string    `json:"status"` // PENDING, IN_PROGRESS, PASSED, FAILED
	Attempts  int       `json:"attempts"`
	LastError string    `json:"last_error,omitempty"`
	Updated   time.Time `json:"updated"`
}

// Status constants
const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusPassed     = "PASSED"
	StatusFailed     = "FAILED"
)

// AgentFramework manages the state of multiple rule implementation agents
type AgentFramework struct {
	StatusDir string
	Agents    map[string]*AgentStatus
}

// NewAgentFramework creates a new agent framework
func NewAgentFramework(statusDir string) *AgentFramework {
	os.MkdirAll(statusDir, 0755)
	return &AgentFramework{
		StatusDir: statusDir,
		Agents:    make(map[string]*AgentStatus),
	}
}

// InitializeAgents creates status entries for all rules
func (af *AgentFramework) InitializeAgents(rules []string) {
	for _, rule := range rules {
		af.Agents[rule] = &AgentStatus{
			Rule:    rule,
			Status:  StatusPending,
			Updated: time.Now(),
		}
		af.saveStatus(rule)
	}
}

// StartAgent marks a rule as in progress
func (af *AgentFramework) StartAgent(rule string) error {
	if _, ok := af.Agents[rule]; !ok {
		return fmt.Errorf("unknown rule: %s", rule)
	}

	af.Agents[rule].Status = StatusInProgress
	af.Agents[rule].Attempts++
	af.Agents[rule].Updated = time.Now()
	af.saveStatus(rule)

	return nil
}

// CompleteAgent marks a rule as passed
func (af *AgentFramework) CompleteAgent(rule string) error {
	af.Agents[rule].Status = StatusPassed
	af.Agents[rule].Updated = time.Now()
	af.saveStatus(rule)

	return nil
}

// FailAgent marks a rule as failed with error message
func (af *AgentFramework) FailAgent(rule, errMsg string) error {
	af.Agents[rule].LastError = errMsg
	af.Agents[rule].Status = StatusFailed
	af.Agents[rule].Updated = time.Now()
	af.saveStatus(rule)

	return nil
}

// GetStatus returns the status of a rule
func (af *AgentFramework) GetStatus(rule string) (*AgentStatus, error) {
	statusFile := filepath.Join(af.StatusDir, rule+".json")
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return nil, err
	}

	var status AgentStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// ListPending returns all rules that are still pending
func (af *AgentFramework) ListPending() []string {
	var pending []string

	for rule, status := range af.Agents {
		if status.Status == StatusPending || status.Status == StatusFailed {
			pending = append(pending, rule)
		}
	}

	return pending
}

// GetReport returns a summary of all agent statuses
func (af *AgentFramework) GetReport() string {
	passed := 0
	failed := 0
	inProgress := 0
	pending := 0

	for _, status := range af.Agents {
		switch status.Status {
		case StatusPassed:
			passed++
		case StatusFailed:
			failed++
		case StatusInProgress:
			inProgress++
		case StatusPending:
			pending++
		}
	}

	return fmt.Sprintf(`Agent Status Report:
  Passed:     %d
  Failed:     %d
  In Progress: %d
  Pending:    %d
  Total:      %d`, passed, failed, inProgress, pending, len(af.Agents))
}

func (af *AgentFramework) saveStatus(rule string) {
	status := af.Agents[rule]
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling status for %s: %v\n", rule, err)
		return
	}

	statusFile := filepath.Join(af.StatusDir, rule+".json")
	os.WriteFile(statusFile, data, 0644)
}
