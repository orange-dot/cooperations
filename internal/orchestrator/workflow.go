package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	ctx "cooperations/internal/context"
	"cooperations/internal/logging"
	"cooperations/internal/tui/stream"
	"cooperations/internal/types"
)

// WorkflowConfig holds workflow execution settings.
type WorkflowConfig struct {
	MaxReviewCycles int               `yaml:"max_review_cycles"`
	RoleTaskTypes   map[string]string `yaml:"role_task_types"`
}

// DefaultWorkflowConfig returns the default workflow configuration.
func DefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		MaxReviewCycles: 2,
	}
}

// emitProgress sends a progress update to the stream if available.
func (o *Orchestrator) emitProgress(stage string, percent float64, message string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Progress <- stream.ProgressUpdate{
		Stage:   stage,
		Percent: percent,
		Message: message,
	}:
	default:
		// Channel full, skip
	}
}

// emitHandoff sends a handoff event to the stream if available.
func (o *Orchestrator) emitHandoff(from, to, reason string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Handoffs <- stream.HandoffEvent{
		From:      from,
		To:        to,
		Reason:    reason,
		Timestamp: time.Now(),
	}:
	default:
	}
}

// emitTokenChunk sends a streamed text chunk if available.
func (o *Orchestrator) emitTokenChunk(role, text string, isFinal bool) {
	if o.stream == nil {
		return
	}
	if text == "" {
		return
	}
	select {
	case o.stream.Tokens <- stream.TokenChunk{
		AgentRole: role,
		Token:     text,
		Timestamp: time.Now(),
		IsFinal:   isFinal,
	}:
	default:
	}
}

// emitThinking sends a thinking update to the stream if available.
func (o *Orchestrator) emitThinking(role, stage string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Thinking <- stream.ThinkingUpdate{
		AgentRole: role,
		Stage:     stage,
		Duration:  0,
	}:
	default:
	}
}

// emitMetrics sends a metrics snapshot to the stream if available.
func (o *Orchestrator) emitMetrics(snapshot stream.MetricsSnapshot) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Metrics <- snapshot:
	default:
	}
}

// emitCode sends a code update to the stream if available.
func (o *Orchestrator) emitCode(path, content, language string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Code <- stream.CodeUpdate{
		Path:     path,
		Content:  content,
		Language: language,
	}:
	default:
	}
}

// emitFileTree sends a file tree update to the stream if available.
func (o *Orchestrator) emitFileTree(action, path string, isDir bool, size int64) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.FileTree <- stream.FileTreeUpdate{
		Action: action,
		Path:   path,
		IsDir:  isDir,
		Size:   size,
	}:
	default:
	}
}

// emitDiff sends a diff update to the stream if available.
func (o *Orchestrator) emitDiff(path, oldContent, newContent string, hunks []stream.DiffHunk) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.FileDiff <- stream.FileDiff{
		Path:       path,
		OldContent: oldContent,
		NewContent: newContent,
		Hunks:      hunks,
	}:
	default:
	}
}

// emitError sends an error to the stream if available.
func (o *Orchestrator) emitError(err error) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Error <- err:
	default:
	}
}

// emitDone signals workflow completion to the stream if available.
func (o *Orchestrator) emitDone() {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Done <- struct{}{}:
	default:
	}
}

// executeWorkflow runs the main workflow loop.
func (o *Orchestrator) executeWorkflow(c context.Context, task types.Task, initialRole types.Role) (types.WorkflowResult, error) {
	state := types.WorkflowState{
		Task:         task,
		Handoffs:     []types.Handoff{},
		CurrentRole:  initialRole,
		ReviewCycles: 0,
	}

	// Create initial handoff context
	handoffCtx := types.HContext{
		TaskDescription: task.Description,
		Requirements:    []string{},
		Constraints:     []string{},
		FilesInScope:    []string{},
	}
	artifacts := types.HArtifacts{}

	// Track total tokens for stream updates
	workflowStart := time.Now()
	totalTokens := 0
	promptTokens := 0
	completionTokens := 0
	stepCount := 0

	// Emit initial progress
	o.emitProgress("Starting", 0, fmt.Sprintf("Starting workflow for task: %s", task.ID))
	o.emitHandoff("user", string(initialRole), "Initial routing")

	// Emit workflow start hook
	startResult := o.hooks.Emit(c, HookEvent{
		Phase:       HookPhaseWorkflowStart,
		TaskID:      task.ID,
		CurrentRole: initialRole,
		Metadata:    map[string]any{"task": task},
	})
	o.emitHookNotify(HookPhaseWorkflowStart, initialRole, task.ID)
	if startResult.Kill {
		return o.abortWorkflow(task, state.Handoffs, startResult.Error)
	}

	for {
		// Check for context cancellation
		select {
		case <-c.Done():
			return o.abortWorkflow(task, state.Handoffs, c.Err())
		default:
		}

		// Get the agent for current role
		agent, ok := o.agents[state.CurrentRole]
		if !ok {
			return o.abortWorkflow(task, state.Handoffs,
				fmt.Errorf("no agent for role: %s", state.CurrentRole))
		}

		// Create handoff for this step
		handoff := ctx.NewHandoff(
			task.ID,
			state.CurrentRole, // from (will be updated after execution)
			state.CurrentRole, // to (current)
			handoffCtx,
			artifacts,
			types.HMetadata{},
		)

		// Pre-agent hook
		preResult := o.hooks.Emit(c, HookEvent{
			Phase:       HookPhasePreAgent,
			TaskID:      task.ID,
			CurrentRole: state.CurrentRole,
			Handoff:     handoff,
		})
		o.emitHookNotify(HookPhasePreAgent, state.CurrentRole, task.ID)

		if preResult.Kill {
			return o.abortWorkflow(task, state.Handoffs, preResult.Error)
		}

		if preResult.Skip {
			// Skip this agent, advance to next
			o.emitProgress("Skipped", float64(stepCount*20),
				fmt.Sprintf("Skipped %s", roleToLabel(state.CurrentRole)))
			o.emitTokenChunk(string(state.CurrentRole),
				fmt.Sprintf("\n[SKIPPED: %s]\n", roleToLabel(state.CurrentRole)), true)

			// Determine next role (use default progression)
			nextRole := o.defaultNextRole(state.CurrentRole)
			if nextRole == nil {
				// Workflow complete
				return o.completeWorkflow(task, state, artifacts)
			}
			o.emitHandoff(string(state.CurrentRole), string(*nextRole), "Skipped to next agent")
			state.CurrentRole = *nextRole
			continue
		}

		if preResult.ModifiedHandoff != nil {
			handoff = preResult.ModifiedHandoff
		}

		// Execute the agent
		logging.AgentStart(string(state.CurrentRole), task.ID)
		stepCount++

		// Emit progress before execution
		roleLabel := roleToLabel(state.CurrentRole)
		o.emitThinking(string(state.CurrentRole), "analyzing")
		o.emitProgress(roleLabel, float64(stepCount*20), fmt.Sprintf("%s is working...", roleLabel))

		response, err := agent.Execute(c, *handoff)
		if err != nil {
			// Check if it was a kill signal
			if o.hooks.IsKilled() {
				return o.abortWorkflow(task, state.Handoffs, fmt.Errorf("workflow killed"))
			}
			logging.Error("agent execution failed", err, "role", state.CurrentRole, "task_id", task.ID)
			o.emitError(err)
			return o.abortWorkflow(task, state.Handoffs, err)
		}

		// Post-agent hook
		postResult := o.hooks.Emit(c, HookEvent{
			Phase:       HookPhasePostAgent,
			TaskID:      task.ID,
			CurrentRole: state.CurrentRole,
			Response:    &response,
			Handoff:     handoff,
		})
		o.emitHookNotify(HookPhasePostAgent, state.CurrentRole, task.ID)

		if postResult.Kill {
			return o.abortWorkflow(task, state.Handoffs, postResult.Error)
		}

		if postResult.ModifiedResponse != nil {
			response = *postResult.ModifiedResponse
		}

		logging.AgentComplete(string(state.CurrentRole), task.ID, response.DurationMS, response.TokensUsed)

		// Emit token and metrics updates
		totalTokens += response.TokensUsed
		promptTokens += response.TokensUsed / 2
		completionTokens += response.TokensUsed - (response.TokensUsed / 2)
		o.emitMetrics(stream.MetricsSnapshot{
			TotalTokens:      totalTokens,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			EstimatedCostUSD: estimateCostUSD(totalTokens),
			ElapsedTime:      time.Since(workflowStart),
			APICallsCount:    stepCount,
			AgentCycles:      stepCount,
			CurrentAgent:     string(state.CurrentRole),
		})
		if strings.TrimSpace(response.Content) != "" {
			separator := ""
			if stepCount > 1 {
				separator = "\n\n"
			}
			header := fmt.Sprintf("[%s]", strings.ToUpper(roleToLabel(state.CurrentRole)))
			o.emitTokenChunk(string(state.CurrentRole), separator+header+"\n"+response.Content+"\n", true)
		}

		// Update handoff with execution metadata
		modelProvider, modelName, profileName := o.modelInfoForRole(state.CurrentRole)
		rvrTaskType := o.roleTaskTypes[state.CurrentRole]
		handoff.Metadata = types.HMetadata{
			TokensUsed:   response.TokensUsed,
			Model:        modelProvider,
			ModelName:    modelName,
			ModelProfile: profileName,
			DurationMS:   response.DurationMS,
			Confidence:   response.Confidence,
			Uncertainty:  response.Uncertainty,
			RVRTaskType:  rvrTaskType,
		}

		// Merge artifacts
		artifacts = ctx.MergeArtifacts(artifacts, response.Artifacts)

		// Update handoff with merged artifacts
		handoff.Artifacts = artifacts

		// Save artifacts to generated folder
		if state.CurrentRole == types.RoleImplementer {
			files := extractFiles(response.Artifacts)
			if len(files) > 0 {
				for rawPath, content := range files {
					cleanPath := cleanRelativePath(rawPath)
					if cleanPath == "" {
						logging.Error("invalid file path in response", nil, "task_id", task.ID, "path", rawPath)
						continue
					}

					oldContent, existed := readFileIfExists(cleanPath)
					path, err := o.store.SaveGeneratedCode(task.ID, cleanPath, content)
					if err != nil {
						logging.Error("failed to save code artifact", err, "task_id", task.ID)
					} else {
						logging.Info("saved code artifact", "path", path)
						o.emitCode(cleanPath, content, detectLanguage(cleanPath))
					}

					if err := writeWorkspaceFile(cleanPath, content); err != nil {
						logging.Error("failed to write task output", err, "task_id", task.ID, "path", cleanPath)
					} else {
						logging.Info("wrote task output", "path", cleanPath)
						if oldContent != content {
							action := "add"
							if existed {
								action = "modify"
							}
							o.emitFileTree(action, cleanPath, false, int64(len(content)))
							o.emitDiff(cleanPath, oldContent, content, simpleDiffHunks(oldContent, content))
						}
					}
				}
			} else if artifacts.Code != "" {
				path, err := o.store.SaveGeneratedCode(task.ID, "", artifacts.Code)
				if err != nil {
					logging.Error("failed to save code artifact", err, "task_id", task.ID)
				} else {
					logging.Info("saved code artifact", "path", path)
					o.emitCode(path, artifacts.Code, "go")
				}

				if targetPath := extractTargetPath(task.Description); targetPath != "" {
					oldContent, existed := readFileIfExists(targetPath)
					if err := writeWorkspaceFile(targetPath, artifacts.Code); err != nil {
						logging.Error("failed to write task output", err, "task_id", task.ID, "path", targetPath)
					} else {
						logging.Info("wrote task output", "path", targetPath)
						o.emitCode(targetPath, artifacts.Code, detectLanguage(targetPath))
						if oldContent != artifacts.Code {
							action := "add"
							if existed {
								action = "modify"
							}
							o.emitFileTree(action, targetPath, false, int64(len(artifacts.Code)))
							o.emitDiff(targetPath, oldContent, artifacts.Code, simpleDiffHunks(oldContent, artifacts.Code))
						}
					}
				}
			}
		}
		if artifacts.DesignDoc != "" && state.CurrentRole == types.RoleArchitect {
			path, err := o.store.SaveDesignDoc(task.ID, artifacts.DesignDoc)
			if err != nil {
				logging.Error("failed to save design doc", err, "task_id", task.ID)
			} else {
				logging.Info("saved design doc", "path", path)
			}
		}
		if artifacts.ReviewFeedback != "" && state.CurrentRole == types.RoleReviewer {
			path, err := o.store.SaveReviewFeedback(task.ID, artifacts.ReviewFeedback)
			if err != nil {
				logging.Error("failed to save review feedback", err, "task_id", task.ID)
			} else {
				logging.Info("saved review feedback", "path", path)
			}
		}

		// Determine next role
		var nextRole *types.Role
		if response.NextRole != nil {
			nextRole = response.NextRole
		}

		// Update handoff with next role
		if nextRole != nil {
			handoff.ToRole = *nextRole
		}

		// Save handoff
		state.Handoffs = append(state.Handoffs, *handoff)
		if err := o.store.SaveHandoff(task.ID, *handoff); err != nil {
			logging.Error("failed to save handoff", err, "task_id", task.ID)
		}

		// Check if workflow is complete
		if nextRole == nil {
			// Workflow end hook
			o.hooks.Emit(c, HookEvent{
				Phase:       HookPhaseWorkflowEnd,
				TaskID:      task.ID,
				CurrentRole: state.CurrentRole,
				Metadata:    map[string]any{"success": true},
			})
			o.emitHookNotify(HookPhaseWorkflowEnd, state.CurrentRole, task.ID)
			return o.completeWorkflow(task, state, artifacts)
		}

		// Pre-handoff hook
		handoffResult := o.hooks.Emit(c, HookEvent{
			Phase:       HookPhasePreHandoff,
			TaskID:      task.ID,
			CurrentRole: state.CurrentRole,
			NextRole:    nextRole,
			Handoff:     handoff,
			Response:    &response,
		})
		o.emitHookNotify(HookPhasePreHandoff, state.CurrentRole, task.ID)

		if handoffResult.Kill {
			return o.abortWorkflow(task, state.Handoffs, handoffResult.Error)
		}

		if handoffResult.Skip {
			// Skip handoff, go to default next
			nextRole = o.defaultNextRole(state.CurrentRole)
			if nextRole == nil {
				return o.completeWorkflow(task, state, artifacts)
			}
		}

		// Check review cycle limit
		if *nextRole == types.RoleReviewer {
			state.ReviewCycles++
			if state.ReviewCycles > o.config.MaxReviewCycles {
				logging.WorkflowComplete(task.ID, false, state.ReviewCycles)
				return types.WorkflowResult{
					Task:      task,
					Handoffs:  state.Handoffs,
					Success:   false,
					Error:     fmt.Sprintf("exceeded max review cycles (%d)", o.config.MaxReviewCycles),
					Artifacts: artifacts,
				}, nil
			}
		}

		// Transition to next role
		logging.Handoff(string(state.CurrentRole), string(*nextRole), task.ID)

		// Emit handoff event
		o.emitHandoff(string(state.CurrentRole), string(*nextRole), fmt.Sprintf("Transitioning to %s", roleToLabel(*nextRole)))

		// Post-handoff hook
		o.hooks.Emit(c, HookEvent{
			Phase:       HookPhasePostHandoff,
			TaskID:      task.ID,
			CurrentRole: *nextRole,
			Handoff:     handoff,
		})
		o.emitHookNotify(HookPhasePostHandoff, *nextRole, task.ID)

		state.CurrentRole = *nextRole

		// Update context for next iteration
		handoffCtx.TaskDescription = response.Content
	}
}

// roleToLabel converts a role to a human-readable label.
func roleToLabel(role types.Role) string {
	switch role {
	case types.RoleArchitect:
		return "Architect"
	case types.RoleImplementer:
		return "Implementer"
	case types.RoleReviewer:
		return "Reviewer"
	case types.RoleNavigator:
		return "Navigator"
	case types.RoleHuman:
		return "Human"
	default:
		return string(role)
	}
}

// modelInfoForRole returns provider, model name, and profile for a role.
func (o *Orchestrator) modelInfoForRole(role types.Role) (string, string, string) {
	profileName := o.roleProfiles[role]
	profile, ok := o.modelProfiles[profileName]
	if !ok {
		return string(types.ModelClaudeCLI), "", profileName
	}
	provider := normalizeProvider(profile.Provider)
	if provider == "" {
		provider = string(types.ModelClaudeCLI)
	}
	switch provider {
	case "codex-cli":
		return provider, profile.Codex.Model, profileName
	default:
		return provider, profile.Claude.Model, profileName
	}
}

// abortWorkflow handles workflow termination.
func (o *Orchestrator) abortWorkflow(task types.Task, handoffs []types.Handoff, err error) (types.WorkflowResult, error) {
	errMsg := "workflow aborted"
	if err != nil {
		errMsg = err.Error()
	}

	o.emitProgress("Aborted", 0, errMsg)
	o.emitError(err)
	o.emitDone()

	return types.WorkflowResult{
		Task:     task,
		Handoffs: handoffs,
		Success:  false,
		Error:    errMsg,
	}, err
}

// completeWorkflow handles successful completion.
func (o *Orchestrator) completeWorkflow(task types.Task, state types.WorkflowState, artifacts types.HArtifacts) (types.WorkflowResult, error) {
	logging.WorkflowComplete(task.ID, true, state.ReviewCycles)
	o.emitProgress("Complete", 100, "Workflow completed successfully")
	o.emitDone()

	task.Status = types.TaskStatusCompleted
	if path, err := o.store.SaveTaskSummary(task.ID, task, artifacts); err != nil {
		logging.Error("failed to save task summary", err, "task_id", task.ID)
	} else {
		logging.Info("saved task summary", "path", path)
	}

	return types.WorkflowResult{
		Task:      task,
		Handoffs:  state.Handoffs,
		Success:   true,
		Artifacts: artifacts,
	}, nil
}

// defaultNextRole returns the default next role in the workflow.
func (o *Orchestrator) defaultNextRole(current types.Role) *types.Role {
	// Default progression: architect -> implementer -> reviewer -> done
	switch current {
	case types.RoleArchitect:
		r := types.RoleImplementer
		return &r
	case types.RoleImplementer:
		r := types.RoleReviewer
		return &r
	case types.RoleReviewer:
		return nil // Done
	case types.RoleNavigator:
		r := types.RoleArchitect
		return &r
	default:
		return nil
	}
}

// emitHookNotify sends a hook notification to the TUI.
func (o *Orchestrator) emitHookNotify(phase HookPhase, role types.Role, taskID string) {
	if o.stream == nil {
		return
	}
	canSkip := phase == HookPhasePreAgent || phase == HookPhasePreHandoff
	select {
	case o.stream.HookNotify <- stream.HookNotification{
		Phase:     stream.HookPhase(phase),
		TaskID:    taskID,
		Role:      string(role),
		Timestamp: time.Now(),
		Paused:    o.hooks.IsPaused(),
		CanSkip:   canSkip,
	}:
	default:
	}
}

func (o *Orchestrator) syncPause(paused *bool) {
	if o.stream == nil || o.stream.Pause == nil {
		return
	}
	for {
		select {
		case v, ok := <-o.stream.Pause:
			if !ok {
				return
			}
			*paused = v
		default:
			return
		}
	}
}

func (o *Orchestrator) waitWhilePaused(ctx context.Context, paused *bool) error {
	if o.stream == nil || o.stream.Pause == nil {
		return nil
	}

	o.syncPause(paused)
	if !*paused {
		return nil
	}

	o.emitProgress("Paused", 0, "Workflow paused")

	for *paused {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v, ok := <-o.stream.Pause:
			if !ok {
				return nil
			}
			*paused = v
		}
	}

	o.emitProgress("Resumed", 0, "Workflow resumed")
	return nil
}

var taskPathPattern = regexp.MustCompile(`(?i)([A-Za-z0-9][A-Za-z0-9_./\\-]*\.[A-Za-z0-9]{1,6})`)

func extractTargetPath(task string) string {
	match := taskPathPattern.FindStringSubmatch(task)
	if len(match) < 2 {
		return ""
	}

	candidate := strings.Trim(match[1], "`\"'()[]{}<>.,;:")
	return cleanRelativePath(candidate)
}

func extractFiles(artifacts map[string]any) map[string]string {
	raw, ok := artifacts["files"]
	if !ok {
		return nil
	}

	switch v := raw.(type) {
	case map[string]string:
		return v
	case map[string]any:
		out := make(map[string]string, len(v))
		for key, value := range v {
			str, ok := value.(string)
			if !ok {
				continue
			}
			out[key] = str
		}
		return out
	default:
		return nil
	}
}

func cleanRelativePath(candidate string) string {
	if candidate == "" {
		return ""
	}

	clean := filepath.Clean(candidate)
	if clean == "." || clean == ".." {
		return ""
	}
	if filepath.IsAbs(clean) {
		return ""
	}
	if strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return ""
	}
	if strings.Contains(clean, ":") {
		return ""
	}

	return clean
}

func detectLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		return "text"
	}
}

func estimateCostUSD(totalTokens int) float64 {
	const costPerMToken = 15.0
	return float64(totalTokens) / 1_000_000 * costPerMToken
}

func readFileIfExists(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func simpleDiffHunks(oldContent, newContent string) []stream.DiffHunk {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	if len(oldLines) == 1 && oldLines[0] == "" {
		oldLines = nil
	}
	if len(newLines) == 1 && newLines[0] == "" {
		newLines = nil
	}

	hunk := stream.DiffHunk{
		OldStart: 1,
		OldCount: len(oldLines),
		NewStart: 1,
		NewCount: len(newLines),
	}
	for _, line := range oldLines {
		hunk.Lines = append(hunk.Lines, stream.DiffLine{Type: "remove", Content: line})
	}
	for _, line := range newLines {
		hunk.Lines = append(hunk.Lines, stream.DiffLine{Type: "add", Content: line})
	}

	if len(hunk.Lines) == 0 {
		return nil
	}
	return []stream.DiffHunk{hunk}
}

func writeWorkspaceFile(relPath string, content string) error {
	dir := filepath.Dir(relPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create target directory: %w", err)
		}
	}

	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write target file: %w", err)
	}

	return nil
}
