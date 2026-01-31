package stream

import (
	"sync"
	"time"
)

// WorkflowStream provides channels for streaming workflow events to TUI.
type WorkflowStream struct {
	// Real-time streaming
	Tokens   chan TokenChunk
	Thinking chan ThinkingUpdate

	// Workflow events
	Progress chan ProgressUpdate
	Handoffs chan HandoffEvent
	AgentLog chan AgentLogEntry

	// Code & files
	Code     chan CodeUpdate
	FileDiff chan FileDiff
	FileTree chan FileTreeUpdate

	// Metrics
	Metrics chan MetricsSnapshot

	// Interaction
	Decision chan DecisionRequest
	Response chan HumanDecision
	Toast    chan ToastNotification

	// Session
	Session chan SessionEvent

	// Control
	Done       chan struct{}
	Error      chan error
	Pause      chan bool
	Control    chan ControlEvent      // Bidirectional control signals
	HookNotify chan HookNotification  // Hook state notifications
	RVR        chan RVREvent          // RVR processing events
	RVRResult  chan RVRResultEvent    // RVR final results

	closeOnce sync.Once
}

// NewWorkflowStream creates a new stream with all channels initialized.
func NewWorkflowStream() *WorkflowStream {
	return &WorkflowStream{
		// High-frequency channels get larger buffers
		Tokens:   make(chan TokenChunk, 100),
		Thinking: make(chan ThinkingUpdate, 10),

		Progress: make(chan ProgressUpdate, 20),
		Handoffs: make(chan HandoffEvent, 10),
		AgentLog: make(chan AgentLogEntry, 50),

		Code:     make(chan CodeUpdate, 10),
		FileDiff: make(chan FileDiff, 10),
		FileTree: make(chan FileTreeUpdate, 20),

		Metrics: make(chan MetricsSnapshot, 10),

		Decision: make(chan DecisionRequest, 1),
		Response: make(chan HumanDecision, 1),
		Toast:    make(chan ToastNotification, 10),

		Session: make(chan SessionEvent, 5),

		Done:       make(chan struct{}),
		Error:      make(chan error, 1),
		Pause:      make(chan bool, 1),
		Control:    make(chan ControlEvent, 10),
		HookNotify: make(chan HookNotification, 20),
		RVR:        make(chan RVREvent, 20),
		RVRResult:  make(chan RVRResultEvent, 5),
	}
}

// Close closes all channels safely.
func (s *WorkflowStream) Close() {
	if s == nil {
		return
	}
	s.closeOnce.Do(func() {
		close(s.Tokens)
		close(s.Thinking)
		close(s.Progress)
		close(s.Handoffs)
		close(s.AgentLog)
		close(s.Code)
		close(s.FileDiff)
		close(s.FileTree)
		close(s.Metrics)
		close(s.Decision)
		close(s.Response)
		close(s.Toast)
		close(s.Session)
		close(s.Done)
		close(s.Error)
		close(s.Pause)
		close(s.Control)
		close(s.HookNotify)
		close(s.RVR)
		close(s.RVRResult)
	})
}

// SendToken sends a token chunk, non-blocking.
func (s *WorkflowStream) SendToken(chunk TokenChunk) {
	select {
	case s.Tokens <- chunk:
	default:
	}
}

// SendProgress sends a progress update, non-blocking.
func (s *WorkflowStream) SendProgress(p ProgressUpdate) {
	select {
	case s.Progress <- p:
	default:
	}
}

// SendHandoff sends a handoff event, non-blocking.
func (s *WorkflowStream) SendHandoff(h HandoffEvent) {
	select {
	case s.Handoffs <- h:
	default:
	}
}

// SendCode sends a code update, non-blocking.
func (s *WorkflowStream) SendCode(c CodeUpdate) {
	select {
	case s.Code <- c:
	default:
	}
}

// SendMetrics sends a metrics snapshot, non-blocking.
func (s *WorkflowStream) SendMetrics(m MetricsSnapshot) {
	select {
	case s.Metrics <- m:
	default:
	}
}

// SendToast sends a toast notification, non-blocking.
func (s *WorkflowStream) SendToast(t ToastNotification) {
	select {
	case s.Toast <- t:
	default:
	}
}

// SendLog sends an agent log entry, non-blocking.
func (s *WorkflowStream) SendLog(l AgentLogEntry) {
	select {
	case s.AgentLog <- l:
	default:
	}
}

// RequestDecision sends a decision request and waits for response.
func (s *WorkflowStream) RequestDecision(d DecisionRequest) HumanDecision {
	s.Decision <- d
	return <-s.Response
}

// SignalDone signals workflow completion.
func (s *WorkflowStream) SignalDone() {
	select {
	case s.Done <- struct{}{}:
	default:
	}
}

// SendError sends an error, non-blocking.
func (s *WorkflowStream) SendError(err error) {
	select {
	case s.Error <- err:
	default:
	}
}

// SendControl sends a control signal to the orchestrator.
func (s *WorkflowStream) SendControl(signal ControlSignal, reason string) {
	select {
	case s.Control <- ControlEvent{
		Signal:    signal,
		Timestamp: time.Now(),
		Reason:    reason,
	}:
	default:
	}
}

// SendHookNotify sends a hook notification to the TUI.
func (s *WorkflowStream) SendHookNotify(n HookNotification) {
	select {
	case s.HookNotify <- n:
	default:
	}
}

// SendRVR sends an RVR processing event.
func (s *WorkflowStream) SendRVR(e RVREvent) {
	select {
	case s.RVR <- e:
	default:
	}
}

// SendRVRResult sends final RVR results.
func (s *WorkflowStream) SendRVRResult(r RVRResultEvent) {
	select {
	case s.RVRResult <- r:
	default:
	}
}
