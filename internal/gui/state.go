// internal/gui/state.go
package gui

import (
	"sync"
	"time"
)

// HandoffEntry tracks a single role handoff event for activity/history display.
type HandoffEntry struct {
	FromRole  string
	ToRole    string
	Timestamp time.Time
}

// WorkflowStepState represents UI state for a single workflow step.
type WorkflowStepState struct {
	ID       string
	Label    string
	Status   string // "pending", "inprogress", "complete", "waiting"
	Progress float32
	Subtext  string
}

// AppState is a thread-safe container for all GUI-relevant state.
type AppState struct {
	mu sync.RWMutex

	// Connection state
	Connected bool

	// Task info
	TaskDescription string
	TaskInProgress  bool
	Completed       bool

	// Required inputs tracking
	RequiredInputs []string
	CurrentInput   int

	// Status display strings
	StatusLine  string
	ActivityLog []string

	// Workflow visualization
	WorkflowSteps []WorkflowStepState
	CurrentStep   int

	// User input UI state
	InputText string

	// Role handoff tracking
	HandoffHistory []HandoffEntry

	// Error state
	ErrorMessage string

	// Prompt indicator
	WaitingForInput bool

	// Completion message
	CompletionMessage string

	// UI state flags
	ShowHelp bool
}

// NewAppState constructs an AppState with sensible defaults.
func NewAppState() *AppState {
	return &AppState{
		Connected:        false,
		TaskDescription:   "",
		TaskInProgress:    false,
		Completed:         false,
		RequiredInputs:    []string{},
		CurrentInput:      0,
		StatusLine:        "",
		ActivityLog:       []string{},
		WorkflowSteps:     []WorkflowStepState{},
		CurrentStep:       0,
		InputText:         "",
		HandoffHistory:    []HandoffEntry{},
		ErrorMessage:      "",
		WaitingForInput:   false,
		CompletionMessage: "",
		ShowHelp:          false,
	}
}

func (s *AppState) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Connected
}

func (s *AppState) SetConnected(connected bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Connected = connected
}

func (s *AppState) SetTaskDescription(desc string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TaskDescription = desc
}

func (s *AppState) SetTaskInProgress(inProgress bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TaskInProgress = inProgress
}

func (s *AppState) SetCompleted(completed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Completed = completed
}

func (s *AppState) SetRequiredInputs(inputs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if inputs == nil {
		s.RequiredInputs = []string{}
		s.CurrentInput = 0
		return
	}

	copied := make([]string, len(inputs))
	copy(copied, inputs)
	s.RequiredInputs = copied

	// Keep CurrentInput within bounds.
	if s.CurrentInput < 0 {
		s.CurrentInput = 0
	}
	if len(s.RequiredInputs) == 0 {
		s.CurrentInput = 0
	} else if s.CurrentInput >= len(s.RequiredInputs) {
		s.CurrentInput = len(s.RequiredInputs) - 1
	}
}

func (s *AppState) SetCurrentInput(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 {
		index = 0
	}
	if n := len(s.RequiredInputs); n == 0 {
		s.CurrentInput = 0
		return
	}
	if index >= len(s.RequiredInputs) {
		index = len(s.RequiredInputs) - 1
	}
	s.CurrentInput = index
}

func (s *AppState) SetStatusLine(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StatusLine = status
}

func (s *AppState) AddActivity(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActivityLog = append(s.ActivityLog, message)
}

func (s *AppState) ClearActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActivityLog = nil
}

func (s *AppState) SetWorkflowSteps(steps []WorkflowStepState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if steps == nil {
		s.WorkflowSteps = []WorkflowStepState{}
		s.CurrentStep = 0
		return
	}

	copied := make([]WorkflowStepState, len(steps))
	copy(copied, steps)
	s.WorkflowSteps = copied

	// Keep CurrentStep within bounds.
	if s.CurrentStep < 0 {
		s.CurrentStep = 0
	}
	if len(s.WorkflowSteps) == 0 {
		s.CurrentStep = 0
	} else if s.CurrentStep >= len(s.WorkflowSteps) {
		s.CurrentStep = len(s.WorkflowSteps) - 1
	}
}

func (s *AppState) SetCurrentStep(step int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if step < 0 {
		step = 0
	}
	if n := len(s.WorkflowSteps); n == 0 {
		s.CurrentStep = 0
		return
	}
	if step >= len(s.WorkflowSteps) {
		step = len(s.WorkflowSteps) - 1
	}
	s.CurrentStep = step
}

func (s *AppState) UpdateStepStatus(stepIndex int, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stepIndex < 0 || stepIndex >= len(s.WorkflowSteps) {
		return
	}
	s.WorkflowSteps[stepIndex].Status = status
}

func (s *AppState) UpdateStepProgress(stepIndex int, progress float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stepIndex < 0 || stepIndex >= len(s.WorkflowSteps) {
		return
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	s.WorkflowSteps[stepIndex].Progress = progress
}

func (s *AppState) UpdateStepSubtext(stepIndex int, subtext string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stepIndex < 0 || stepIndex >= len(s.WorkflowSteps) {
		return
	}
	s.WorkflowSteps[stepIndex].Subtext = subtext
}

func (s *AppState) SetInputText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InputText = text
}

func (s *AppState) GetInputText() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.InputText
}

func (s *AppState) AddHandoff(fromRole, toRole string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.HandoffHistory = append(s.HandoffHistory, HandoffEntry{
		FromRole:  fromRole,
		ToRole:    toRole,
		Timestamp: time.Now(),
	})
}

func (s *AppState) ClearHandoffHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HandoffHistory = nil
}

func (s *AppState) SetError(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ErrorMessage = message
}

func (s *AppState) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ErrorMessage = ""
}

func (s *AppState) SetWaitingForInput(waiting bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.WaitingForInput = waiting
}

func (s *AppState) SetCompletionMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CompletionMessage = message
}

func (s *AppState) ToggleHelp() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ShowHelp = !s.ShowHelp
}

// StateSnapshot is a copy of AppState without the mutex, safe to pass by value.
type StateSnapshot struct {
	Connected         bool
	TaskDescription   string
	TaskInProgress    bool
	Completed         bool
	RequiredInputs    []string
	CurrentInput      int
	StatusLine        string
	ActivityLog       []string
	WorkflowSteps     []WorkflowStepState
	CurrentStep       int
	InputText         string
	HandoffHistory    []HandoffEntry
	ErrorMessage      string
	WaitingForInput   bool
	CompletionMessage string
	ShowHelp          bool
}

func (s *AppState) Snapshot() StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := StateSnapshot{
		Connected:         s.Connected,
		TaskDescription:   s.TaskDescription,
		TaskInProgress:    s.TaskInProgress,
		Completed:         s.Completed,
		CurrentInput:      s.CurrentInput,
		StatusLine:        s.StatusLine,
		CurrentStep:       s.CurrentStep,
		InputText:         s.InputText,
		ErrorMessage:      s.ErrorMessage,
		WaitingForInput:   s.WaitingForInput,
		CompletionMessage: s.CompletionMessage,
		ShowHelp:          s.ShowHelp,
	}

	if s.RequiredInputs != nil {
		cp.RequiredInputs = make([]string, len(s.RequiredInputs))
		copy(cp.RequiredInputs, s.RequiredInputs)
	} else {
		cp.RequiredInputs = []string{}
	}

	if s.ActivityLog != nil {
		cp.ActivityLog = make([]string, len(s.ActivityLog))
		copy(cp.ActivityLog, s.ActivityLog)
	} else {
		cp.ActivityLog = []string{}
	}

	if s.WorkflowSteps != nil {
		cp.WorkflowSteps = make([]WorkflowStepState, len(s.WorkflowSteps))
		copy(cp.WorkflowSteps, s.WorkflowSteps)
	} else {
		cp.WorkflowSteps = []WorkflowStepState{}
	}

	if s.HandoffHistory != nil {
		cp.HandoffHistory = make([]HandoffEntry, len(s.HandoffHistory))
		copy(cp.HandoffHistory, s.HandoffHistory)
	} else {
		cp.HandoffHistory = []HandoffEntry{}
	}

	return cp
}