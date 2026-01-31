package orchestrator

import (
	"context"
	"fmt"
	"os"

	"cooperations/internal/adapters"
	"cooperations/internal/agents"
	coopctx "cooperations/internal/context"
	"cooperations/internal/rvr"
	"cooperations/internal/tui/stream"
	"cooperations/internal/types"
)

// Orchestrator coordinates agents to complete tasks.
type Orchestrator struct {
	router        *Router
	agents        map[types.Role]agents.Agent
	store         *coopctx.Store
	config        WorkflowConfig
	stream        *stream.WorkflowStream // Optional stream for GUI events
	roleProfiles  map[types.Role]string
	modelProfiles map[string]ModelProfile
	roleTaskTypes map[types.Role]string
	rvrConfig     *rvr.RVRConfig
	hooks         *HookController // Hook controller for workflow control
}

// New creates a new orchestrator with the given configuration.
func New(config WorkflowConfig) (*Orchestrator, error) {
	appCfg := DefaultAppConfig()
	appCfg.Workflow.MaxReviewCycles = config.MaxReviewCycles
	if len(config.RoleTaskTypes) > 0 {
		appCfg.Workflow.RoleTaskTypes = config.RoleTaskTypes
	}
	return NewFromConfig(appCfg)
}

// NewFromConfig creates a new orchestrator from full app config.
func NewFromConfig(cfg AppConfig) (*Orchestrator, error) {
	cfg = ApplyAppDefaults(cfg)
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

	// Normalize role mappings and task types
	roleProfiles, err := normalizeRoleProfiles(cfg.Roles)
	if err != nil {
		return nil, err
	}
	roleTaskTypes, err := normalizeRoleTaskTypes(cfg.Workflow.RoleTaskTypes)
	if err != nil {
		return nil, err
	}

	// Initialize CLIs per profile
	cliCache := make(map[string]adapters.CLI)
	agentMap := make(map[types.Role]agents.Agent)

	for role, profileName := range roleProfiles {
		profile, ok := cfg.Models[profileName]
		if !ok {
			return nil, fmt.Errorf("model profile not found: %s", profileName)
		}
		provider := normalizeProvider(profile.Provider)
		if provider == "" {
			return nil, fmt.Errorf("model profile %s missing provider", profileName)
		}

		cli, ok := cliCache[profileName]
		if !ok {
			switch provider {
			case "claude-cli":
				created, err := adapters.NewClaudeCLIWithConfig(profile.Claude)
				if err != nil {
					return nil, fmt.Errorf("create claude CLI: %w", err)
				}
				cli = created
			case "codex-cli":
				created, err := adapters.NewCodexCLIWithConfig(repoRoot, profile.Codex)
				if err != nil {
					return nil, fmt.Errorf("create codex CLI: %w", err)
				}
				cli = created
			default:
				return nil, fmt.Errorf("unsupported provider: %s", provider)
			}
			cliCache[profileName] = cli
		}

		taskType := roleTaskTypes[role]
		switch role {
		case types.RoleArchitect:
			agentMap[role] = agents.NewArchitectAgent(cli, &cfg.RVR, taskType)
		case types.RoleImplementer:
			agentMap[role] = agents.NewImplementerAgent(cli, &cfg.RVR, taskType)
		case types.RoleReviewer:
			agentMap[role] = agents.NewReviewerAgent(cli, &cfg.RVR, taskType)
		case types.RoleNavigator:
			agentMap[role] = agents.NewNavigatorAgent(cli, &cfg.RVR, taskType)
		case types.RoleHuman:
			// Human agent not configured here
		}
	}

	return &Orchestrator{
		router:        NewRouter(),
		agents:        agentMap,
		store:         store,
		config:        cfg.Workflow,
		stream:        nil,
		roleProfiles:  roleProfiles,
		modelProfiles: cfg.Models,
		roleTaskTypes: roleTaskTypes,
		rvrConfig:     &cfg.RVR,
		hooks:         NewHookController(),
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

// NewWithStreamFromConfig creates a new orchestrator that emits events to the given stream.
func NewWithStreamFromConfig(cfg AppConfig, ws *stream.WorkflowStream) (*Orchestrator, error) {
	orch, err := NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	orch.stream = ws
	return orch, nil
}

// Run executes a task through the workflow.
func (o *Orchestrator) Run(ctx context.Context, taskDescription string) (types.WorkflowResult, error) {
	// Reset hook controller for new workflow
	o.hooks.Reset()

	// Start control listener if stream is available
	o.startControlListener(ctx)

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

// Hooks returns the hook controller for external registration.
func (o *Orchestrator) Hooks() *HookController {
	return o.hooks
}

// startControlListener starts a goroutine that forwards stream controls to hooks.
func (o *Orchestrator) startControlListener(ctx context.Context) {
	if o.stream == nil || o.hooks == nil {
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ctrl, ok := <-o.stream.Control:
				if !ok {
					return
				}
				// Convert stream control signal to hook signal
				var sig ControlSignal
				switch ctrl.Signal {
				case stream.ControlStep:
					sig = SignalStep
				case stream.ControlSkip:
					sig = SignalSkip
				case stream.ControlKill:
					sig = SignalKill
				case stream.ControlPause:
					sig = SignalPause
				case stream.ControlResume:
					sig = SignalResume
				}
				if sig != SignalNone {
					o.hooks.SendSignal(sig)
				}
			}
		}
	}()
}
