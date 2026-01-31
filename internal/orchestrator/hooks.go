// Package orchestrator provides the workflow orchestration layer.
package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"cooperations/internal/types"
)

// HookPhase identifies when a hook runs in the workflow.
type HookPhase string

const (
	HookPhaseWorkflowStart HookPhase = "workflow_start"
	HookPhasePreAgent      HookPhase = "pre_agent"
	HookPhaseMidAgent      HookPhase = "mid_agent"
	HookPhasePostAgent     HookPhase = "post_agent"
	HookPhasePreHandoff    HookPhase = "pre_handoff"
	HookPhasePostHandoff   HookPhase = "post_handoff"
	HookPhaseWorkflowEnd   HookPhase = "workflow_end"
)

// ControlSignal represents a control action from the UI.
type ControlSignal string

const (
	SignalNone   ControlSignal = ""
	SignalStep   ControlSignal = "step"   // Execute one agent, then pause
	SignalSkip   ControlSignal = "skip"   // Skip current agent, go to next
	SignalKill   ControlSignal = "kill"   // Immediate abort with cleanup
	SignalPause  ControlSignal = "pause"  // Pause at next hook point
	SignalResume ControlSignal = "resume" // Resume execution
)

// HookEvent contains context for hook execution.
type HookEvent struct {
	Phase       HookPhase
	TaskID      string
	CurrentRole types.Role
	NextRole    *types.Role
	Handoff     *types.Handoff
	Response    *types.AgentResponse
	Error       error
	Metadata    map[string]any
}

// HookResult determines how the workflow proceeds.
type HookResult struct {
	Continue         bool                  // If false, workflow pauses
	Skip             bool                  // If true, skip current agent
	Kill             bool                  // If true, abort workflow
	ModifiedHandoff  *types.Handoff        // Optional: modified handoff
	ModifiedResponse *types.AgentResponse  // Optional: modified response
	Error            error
}

// Hook is a function called at specific workflow phases.
type Hook func(ctx context.Context, event HookEvent) HookResult

// HookRegistration tracks a registered hook.
type HookRegistration struct {
	ID       string
	Phase    HookPhase
	Priority int // Lower = earlier execution
	Handler  Hook
}

// HookController manages hook registration and execution.
type HookController struct {
	mu     sync.RWMutex
	hooks  map[HookPhase][]HookRegistration
	nextID int

	// Control channels
	controlCh chan ControlSignal

	// State
	paused   bool
	stepMode bool // Auto-pause after each agent
	killed   bool

	// Configuration
	autoStepPause bool // If true, step mode pauses after each agent
}

// NewHookController creates a new hook controller.
func NewHookController() *HookController {
	return &HookController{
		hooks:         make(map[HookPhase][]HookRegistration),
		controlCh:     make(chan ControlSignal, 10),
		autoStepPause: true,
	}
}

// Register adds a hook for a phase.
func (hc *HookController) Register(phase HookPhase, priority int, handler Hook) string {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.nextID++
	id := fmt.Sprintf("hook_%d", hc.nextID)

	reg := HookRegistration{
		ID:       id,
		Phase:    phase,
		Priority: priority,
		Handler:  handler,
	}

	hc.hooks[phase] = append(hc.hooks[phase], reg)
	// Sort by priority
	sort.Slice(hc.hooks[phase], func(i, j int) bool {
		return hc.hooks[phase][i].Priority < hc.hooks[phase][j].Priority
	})

	return id
}

// Unregister removes a hook by ID.
func (hc *HookController) Unregister(id string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for phase, regs := range hc.hooks {
		for i, reg := range regs {
			if reg.ID == id {
				hc.hooks[phase] = append(regs[:i], regs[i+1:]...)
				return
			}
		}
	}
}

// Emit fires hooks for a phase and handles control signals.
func (hc *HookController) Emit(ctx context.Context, event HookEvent) HookResult {
	// Check for pending control signals first (non-blocking)
	select {
	case sig := <-hc.controlCh:
		result := hc.handleSignal(sig)
		if !result.Continue || result.Skip || result.Kill {
			return result
		}
	default:
	}

	// Check if killed
	hc.mu.RLock()
	killed := hc.killed
	paused := hc.paused
	hc.mu.RUnlock()

	if killed {
		return HookResult{Kill: true, Error: fmt.Errorf("workflow killed")}
	}

	// Check if paused - wait for resume
	if paused {
		result := hc.waitForResume(ctx)
		if !result.Continue || result.Skip || result.Kill {
			return result
		}
	}

	// Execute hooks for this phase
	hc.mu.RLock()
	hooks := make([]HookRegistration, len(hc.hooks[event.Phase]))
	copy(hooks, hc.hooks[event.Phase])
	hc.mu.RUnlock()

	var lastResult HookResult
	lastResult.Continue = true

	for _, reg := range hooks {
		result := reg.Handler(ctx, event)
		if !result.Continue || result.Skip || result.Kill {
			return result
		}
		if result.ModifiedHandoff != nil {
			lastResult.ModifiedHandoff = result.ModifiedHandoff
		}
		if result.ModifiedResponse != nil {
			lastResult.ModifiedResponse = result.ModifiedResponse
		}
	}

	// Check step mode - auto-pause after post-agent phase
	hc.mu.Lock()
	if hc.stepMode && hc.autoStepPause && event.Phase == HookPhasePostAgent {
		hc.paused = true
		hc.stepMode = false
	}
	hc.mu.Unlock()

	return lastResult
}

// SendSignal sends a control signal and updates state immediately.
func (hc *HookController) SendSignal(sig ControlSignal) {
	// Update state immediately for signals that change state
	hc.mu.Lock()
	switch sig {
	case SignalPause:
		hc.paused = true
	case SignalResume:
		hc.paused = false
	case SignalKill:
		hc.killed = true
	case SignalStep:
		hc.stepMode = true
		hc.paused = false
	}
	hc.mu.Unlock()

	// Also send to channel for workflow to pick up
	select {
	case hc.controlCh <- sig:
	default:
		// Channel full, drain and send
		select {
		case <-hc.controlCh:
		default:
		}
		hc.controlCh <- sig
	}
}

// handleSignal processes a control signal.
func (hc *HookController) handleSignal(sig ControlSignal) HookResult {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	switch sig {
	case SignalStep:
		hc.stepMode = true
		hc.paused = false
		return HookResult{Continue: true}
	case SignalSkip:
		return HookResult{Continue: true, Skip: true}
	case SignalKill:
		hc.killed = true
		return HookResult{Kill: true, Error: fmt.Errorf("workflow killed by user")}
	case SignalPause:
		hc.paused = true
		return HookResult{Continue: false}
	case SignalResume:
		hc.paused = false
		return HookResult{Continue: true}
	}
	return HookResult{Continue: true}
}

// waitForResume blocks until resumed, stepped, skipped, or killed.
func (hc *HookController) waitForResume(ctx context.Context) HookResult {
	for {
		hc.mu.RLock()
		paused := hc.paused
		killed := hc.killed
		hc.mu.RUnlock()

		if !paused || killed {
			break
		}

		select {
		case <-ctx.Done():
			return HookResult{Kill: true, Error: ctx.Err()}
		case sig := <-hc.controlCh:
			result := hc.handleSignal(sig)
			if sig == SignalResume || sig == SignalStep {
				return HookResult{Continue: true}
			}
			if !result.Continue || result.Skip || result.Kill {
				return result
			}
		}
	}

	hc.mu.RLock()
	killed := hc.killed
	hc.mu.RUnlock()

	if killed {
		return HookResult{Kill: true, Error: fmt.Errorf("workflow killed")}
	}
	return HookResult{Continue: true}
}

// SetAutoStepPause configures whether step mode auto-pauses.
func (hc *HookController) SetAutoStepPause(enabled bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.autoStepPause = enabled
}

// IsPaused returns current pause state.
func (hc *HookController) IsPaused() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.paused
}

// IsKilled returns if workflow was killed.
func (hc *HookController) IsKilled() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.killed
}

// SetPaused sets the pause state directly.
func (hc *HookController) SetPaused(paused bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.paused = paused
}

// Reset resets the controller state for a new workflow.
func (hc *HookController) Reset() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.paused = false
	hc.stepMode = false
	hc.killed = false
	// Drain control channel
	for {
		select {
		case <-hc.controlCh:
		default:
			return
		}
	}
}
