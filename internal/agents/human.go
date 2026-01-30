package agents

import (
	"context"
	"fmt"

	"cooperations/internal/gui/stream"
	"cooperations/internal/types"
)

// HumanAgent represents a human-in-the-loop agent.
// It sends handoffs to the UI and waits for a HumanDecision response.
type HumanAgent struct {
	outputChan chan<- types.Handoff        // Send handoffs TO the GUI
	inputChan  <-chan stream.HumanDecision // Receive decisions FROM the GUI
}

// NewHumanAgent creates a new human agent with the given channels.
func NewHumanAgent(output chan<- types.Handoff, input <-chan stream.HumanDecision) *HumanAgent {
	return &HumanAgent{
		outputChan: output,
		inputChan:  input,
	}
}

// Role returns the human role identifier.
func (a *HumanAgent) Role() types.Role {
	return types.RoleHuman
}

// Execute sends the handoff to the GUI and waits for a human decision.
func (a *HumanAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
	// Send handoff to GUI for display
	select {
	case <-ctx.Done():
		return types.AgentResponse{}, ctx.Err()
	case a.outputChan <- handoff:
	}

	// Wait for human decision
	var decision stream.HumanDecision
	select {
	case <-ctx.Done():
		return types.AgentResponse{}, ctx.Err()
	case d, ok := <-a.inputChan:
		if !ok {
			return types.AgentResponse{}, fmt.Errorf("human decision channel closed")
		}
		decision = d
	}

	// Convert decision to AgentResponse
	return a.processDecision(decision, handoff)
}

func (a *HumanAgent) processDecision(decision stream.HumanDecision, handoff types.Handoff) (types.AgentResponse, error) {
	switch decision.Action {
	case stream.DecisionActionApprove:
		// Workflow complete
		return types.AgentResponse{
			Content:   "Human approved the code",
			Artifacts: map[string]any{"human_decision": "approved"},
			NextRole:  nil,
		}, nil

	case stream.DecisionActionReject:
		// Route back to implementer with feedback
		implementer := types.RoleImplementer
		return types.AgentResponse{
			Content:   decision.Comment,
			Artifacts: map[string]any{"review_feedback": decision.Comment},
			NextRole:  &implementer,
		}, nil

	case stream.DecisionActionEdit:
		// Human edited code directly, workflow complete
		return types.AgentResponse{
			Content:   "Human edited the code",
			Artifacts: map[string]any{"code": decision.Edited},
			NextRole:  nil,
		}, nil

	default:
		return types.AgentResponse{}, fmt.Errorf("unknown human action: %v", decision.Action)
	}
}
