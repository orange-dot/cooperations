package orchestrator

import (
	"context"
	"fmt"
	"os"

	"cooperations/internal/adapters"
	"cooperations/internal/agents"
	coopctx "cooperations/internal/context"
	"cooperations/internal/gui/stream"
	"cooperations/internal/types"
)

// Orchestrator coordinates agents to complete tasks.
type Orchestrator struct {
	router *Router
	agents map[types.Role]agents.Agent
	store  *coopctx.Store
	config WorkflowConfig
	stream *stream.WorkflowStream // Optional stream for GUI events
}

// New creates a new orchestrator with the given configuration.
func New(config WorkflowConfig) (*Orchestrator, error) {
	// Initialize store
	storeDir := os.Getenv("COOPERATIONS_DIR")
	if storeDir == "" {
		storeDir = ".cooperations"
	}
	generatedDir := os.Getenv("COOPERATIONS_GENERATED_DIR")
	if generatedDir == "" {
		generatedDir = "generated"
	}

	store, err := coopctx.NewStore(storeDir, generatedDir)
	if err != nil {
		return nil, fmt.Errorf("create store: %w", err)
	}

	// Get repository root directory for Codex
	repoRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	// Initialize CLIs
	claudeCLI, err := adapters.NewClaudeCLI()
	if err != nil {
		return nil, fmt.Errorf("create claude CLI: %w", err)
	}

	codexCLI, err := adapters.NewCodexCLI(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("create codex CLI: %w", err)
	}

	// Initialize agents with their respective CLIs
	agentMap := map[types.Role]agents.Agent{
		types.RoleArchitect:   agents.NewArchitectAgent(claudeCLI),
		types.RoleImplementer: agents.NewImplementerAgent(codexCLI),
		types.RoleReviewer:    agents.NewReviewerAgent(claudeCLI),
		types.RoleNavigator:   agents.NewNavigatorAgent(claudeCLI),
	}

	return &Orchestrator{
		router: NewRouter(),
		agents: agentMap,
		store:  store,
		config: config,
		stream: nil,
	}, nil
}

// NewWithStream creates a new orchestrator that emits events to the given stream.
func NewWithStream(config WorkflowConfig, ws *stream.WorkflowStream) (*Orchestrator, error) {
	orch, err := New(config)
	if err != nil {
		return nil, err
	}
	orch.stream = ws
	return orch, nil
}

// Run executes a task through the workflow.
func (o *Orchestrator) Run(ctx context.Context, taskDescription string) (types.WorkflowResult, error) {
	// Create task
	task, err := o.store.CreateTask(taskDescription)
	if err != nil {
		return types.WorkflowResult{}, fmt.Errorf("create task: %w", err)
	}

	// Update task status
	if err := o.store.UpdateTaskStatus(task.ID, types.TaskStatusInProgress); err != nil {
		return types.WorkflowResult{}, fmt.Errorf("update task status: %w", err)
	}

	// Route to initial role
	initialRole := o.router.Route(taskDescription)

	// Execute workflow
	result, err := o.executeWorkflow(ctx, task, initialRole)

	// Update final task status
	finalStatus := types.TaskStatusCompleted
	if err != nil || !result.Success {
		finalStatus = types.TaskStatusFailed
	}
	if updateErr := o.store.UpdateTaskStatus(task.ID, finalStatus); updateErr != nil {
		// Log but don't fail
		fmt.Fprintf(os.Stderr, "warning: failed to update task status: %v\n", updateErr)
	}

	return result, err
}

// DryRun shows the routing decision without executing.
func (o *Orchestrator) DryRun(taskDescription string) (types.Role, float64) {
	return o.router.RouteWithConfidence(taskDescription)
}

// GetTask retrieves a task by ID.
func (o *Orchestrator) GetTask(id string) (*types.Task, error) {
	return o.store.GetTask(id)
}

// ListTasks returns all tasks.
func (o *Orchestrator) ListTasks() ([]types.Task, error) {
	return o.store.LoadTasks()
}

// GetHandoffs returns handoffs for a task.
func (o *Orchestrator) GetHandoffs(taskID string) ([]types.Handoff, error) {
	return o.store.LoadHandoffs(taskID)
}
