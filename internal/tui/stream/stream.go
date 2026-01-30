package stream

import "sync"

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
	Done  chan struct{}
	Error chan error
	Pause chan bool

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

		Done:  make(chan struct{}),
		Error: make(chan error, 1),
		Pause: make(chan bool, 1),
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
