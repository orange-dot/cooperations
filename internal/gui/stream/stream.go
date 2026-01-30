package stream

import (
	"sync"
)

// WorkflowStream provides a set of channels for streaming workflow events to a GUI.
// All channels are buffered as specified to prevent routine producers from blocking.
type WorkflowStream struct {
	Progress chan ProgressUpdate
	Code     chan CodeUpdate
	Handoffs chan HandoffEvent
	Tokens   chan TokenUpdate

	Decision chan DecisionRequest
	Response chan HumanDecision

	Done  chan struct{}
	Error chan error

	closeOnce sync.Once
}

// NewWorkflowStream constructs a new WorkflowStream with all channels initialized
// with the required buffer sizes.
func NewWorkflowStream() *WorkflowStream {
	return &WorkflowStream{
		Progress: make(chan ProgressUpdate, 10),
		Code:     make(chan CodeUpdate, 10),
		Handoffs: make(chan HandoffEvent, 10),
		Tokens:   make(chan TokenUpdate, 10),

		Decision: make(chan DecisionRequest, 1),
		Response: make(chan HumanDecision, 1),

		Done:  make(chan struct{}),
		Error: make(chan error, 1),
	}
}

// Close closes all channels safely. It is idempotent.
func (s *WorkflowStream) Close() {
	if s == nil {
		return
	}

	s.closeOnce.Do(func() {
		// Order doesn't matter; close all channels exactly once.
		close(s.Progress)
		close(s.Code)
		close(s.Handoffs)
		close(s.Tokens)

		close(s.Decision)
		close(s.Response)

		close(s.Done)
		close(s.Error)
	})
}