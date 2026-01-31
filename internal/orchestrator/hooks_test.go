package orchestrator

import (
	"context"
	"testing"
	"time"

	"cooperations/internal/types"
)

func TestHookController_Register(t *testing.T) {
	hc := NewHookController()

	id := hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		return HookResult{Continue: true}
	})

	if id == "" {
		t.Error("expected non-empty hook ID")
	}
}

func TestHookController_RegisterMultiple(t *testing.T) {
	hc := NewHookController()

	id1 := hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		return HookResult{Continue: true}
	})
	id2 := hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		return HookResult{Continue: true}
	})

	if id1 == id2 {
		t.Error("expected unique hook IDs")
	}
}

func TestHookController_Unregister(t *testing.T) {
	hc := NewHookController()
	called := false

	id := hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		called = true
		return HookResult{Continue: true}
	})

	hc.Unregister(id)
	hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})

	if called {
		t.Error("hook should not have been called after unregister")
	}
}

func TestHookController_Emit(t *testing.T) {
	hc := NewHookController()
	called := false

	hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		called = true
		return HookResult{Continue: true}
	})

	result := hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})

	if !called {
		t.Error("hook was not called")
	}
	if !result.Continue {
		t.Error("expected continue=true")
	}
}

func TestHookController_EmitPriority(t *testing.T) {
	hc := NewHookController()
	order := []int{}

	hc.Register(HookPhasePreAgent, 2, func(ctx context.Context, e HookEvent) HookResult {
		order = append(order, 2)
		return HookResult{Continue: true}
	})
	hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		order = append(order, 0)
		return HookResult{Continue: true}
	})
	hc.Register(HookPhasePreAgent, 1, func(ctx context.Context, e HookEvent) HookResult {
		order = append(order, 1)
		return HookResult{Continue: true}
	})

	hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})

	if len(order) != 3 || order[0] != 0 || order[1] != 1 || order[2] != 2 {
		t.Errorf("hooks not called in priority order: got %v", order)
	}
}

func TestHookController_Skip(t *testing.T) {
	hc := NewHookController()

	// Send skip signal before emit
	hc.SendSignal(SignalSkip)

	result := hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})

	if !result.Skip {
		t.Error("expected skip=true after skip signal")
	}
}

func TestHookController_Kill(t *testing.T) {
	hc := NewHookController()

	hc.SendSignal(SignalKill)

	result := hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})

	if !result.Kill {
		t.Error("expected kill=true after kill signal")
	}
	if !hc.IsKilled() {
		t.Error("expected IsKilled()=true")
	}
}

func TestHookController_Pause(t *testing.T) {
	hc := NewHookController()

	hc.SendSignal(SignalPause)

	if !hc.IsPaused() {
		t.Error("expected IsPaused()=true after pause signal")
	}
}

func TestHookController_PauseResume(t *testing.T) {
	hc := NewHookController()

	hc.SetPaused(true)
	if !hc.IsPaused() {
		t.Error("expected paused=true")
	}

	// Resume in background
	go func() {
		time.Sleep(10 * time.Millisecond)
		hc.SendSignal(SignalResume)
	}()

	result := hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})

	if !result.Continue {
		t.Error("expected continue=true after resume")
	}
	if hc.IsPaused() {
		t.Error("expected paused=false after resume")
	}
}

func TestHookController_StepMode(t *testing.T) {
	hc := NewHookController()
	hc.SetAutoStepPause(true)

	// Send step signal
	hc.SendSignal(SignalStep)

	// Pre-agent should continue
	result1 := hc.Emit(context.Background(), HookEvent{Phase: HookPhasePreAgent})
	if !result1.Continue {
		t.Error("expected continue after step signal at pre-agent")
	}

	// Post-agent should trigger auto-pause
	hc.Emit(context.Background(), HookEvent{Phase: HookPhasePostAgent})

	if !hc.IsPaused() {
		t.Error("expected paused after post-agent in step mode")
	}
}

func TestHookController_Reset(t *testing.T) {
	hc := NewHookController()

	hc.SetPaused(true)
	hc.SendSignal(SignalKill)

	hc.Reset()

	if hc.IsPaused() {
		t.Error("expected paused=false after reset")
	}
	if hc.IsKilled() {
		t.Error("expected killed=false after reset")
	}
}

func TestHookController_ContextCancel(t *testing.T) {
	hc := NewHookController()
	hc.SetPaused(true)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	result := hc.Emit(ctx, HookEvent{Phase: HookPhasePreAgent})

	if !result.Kill {
		t.Error("expected kill=true after context cancel")
	}
}

func TestHookController_ModifyHandoff(t *testing.T) {
	hc := NewHookController()

	hc.Register(HookPhasePreAgent, 0, func(ctx context.Context, e HookEvent) HookResult {
		// Modify the handoff
		if e.Handoff != nil {
			modified := *e.Handoff
			modified.TaskID = "modified"
			return HookResult{Continue: true, ModifiedHandoff: &modified}
		}
		return HookResult{Continue: true}
	})

	handoff := &types.Handoff{TaskID: "original"}
	result := hc.Emit(context.Background(), HookEvent{
		Phase:   HookPhasePreAgent,
		Handoff: handoff,
	})

	if result.ModifiedHandoff == nil {
		t.Error("expected modified handoff")
	}
	if result.ModifiedHandoff.TaskID != "modified" {
		t.Errorf("expected TaskID='modified', got '%s'", result.ModifiedHandoff.TaskID)
	}
}
