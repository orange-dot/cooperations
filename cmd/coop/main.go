// Package main provides the CLI entry point for the cooperations orchestrator.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"cooperations/internal/gui"
	"cooperations/internal/logging"
	"cooperations/internal/orchestrator"
	"cooperations/internal/tui"
	"cooperations/internal/tui/demo"
	"cooperations/internal/tui/stream"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	verbose      bool
	dryRun       bool
	maxCycles    int
	workflowType string
	outputFile   string
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	// Setup logging
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logging.Setup(logLevel)

	// Root command
	rootCmd := &cobra.Command{
		Use:   "coop",
		Short: "Cooperations - AI mob programming orchestrator",
		Long:  "Coordinates Claude Opus 4.5 and Codex 5.2 as collaborative mob programmers.",
	}

	// Run command
	runCmd := &cobra.Command{
		Use:   "run <task>",
		Short: "Execute a task through the mob programming workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  runTask,
	}
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show routing decision without executing")
	runCmd.Flags().IntVar(&maxCycles, "max-cycles", 0, "Override max review cycles")
	runCmd.Flags().StringVar(&workflowType, "workflow", "", "Force workflow type (feature, bugfix, review)")
	runCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write generated code to file")

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status [task_id]",
		Short: "Show task status",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showStatus,
	}

	// History command
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "List past tasks",
		RunE:  showHistory,
	}
	var historyLimit int
	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Number of tasks to show")

	// GUI command
	guiCmd := &cobra.Command{
		Use:   "gui <task>",
		Short: "Launch the graphical interface for a task",
		Long:  "Opens the futuristic Gio-based GUI for interactive mob programming workflow.",
		Args:  cobra.ExactArgs(1),
		RunE:  runGUI,
	}
	var guiDemoMode bool
	guiCmd.Flags().BoolVar(&guiDemoMode, "demo", false, "Run in demo mode with stub progress")

	// TUI command
	tuiCmd := &cobra.Command{
		Use:   "tui [task]",
		Short: "Launch the terminal user interface",
		Long:  "Opens the Bubble Tea-based TUI for interactive mob programming workflow.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTUI,
	}
	var tuiDemoMode bool
	tuiCmd.Flags().BoolVar(&tuiDemoMode, "demo", false, "Run in demo mode with simulated workflow")

	rootCmd.AddCommand(runCmd, statusCmd, historyCmd, guiCmd, tuiCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTask(cmd *cobra.Command, args []string) error {
	task := args[0]

	// Get max cycles from env or flag
	cycles := 2
	if envCycles := os.Getenv("MAX_REVIEW_CYCLES"); envCycles != "" {
		if c, err := strconv.Atoi(envCycles); err == nil {
			cycles = c
		}
	}
	if maxCycles > 0 {
		cycles = maxCycles
	}

	config := orchestrator.WorkflowConfig{
		MaxReviewCycles: cycles,
	}

	orch, err := orchestrator.New(config)
	if err != nil {
		return fmt.Errorf("initialize orchestrator: %w", err)
	}

	// Dry run mode
	if dryRun {
		role, confidence := orch.DryRun(task)
		fmt.Printf("[DRY-RUN] Task would be routed to: %s (confidence: %.0f%%)\n", role, confidence*100)
		return nil
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, cancelling...")
		cancel()
	}()

	// Run the task
	fmt.Printf("[START] Running task: %s\n", truncate(task, 60))
	result, err := orch.Run(ctx, task)
	if err != nil {
		return fmt.Errorf("run task: %w", err)
	}

	// Print result
	if result.Success {
		fmt.Printf("[COMPLETE] Task %s completed successfully\n", result.Task.ID)
	} else {
		fmt.Printf("[FAILED] Task %s failed: %s\n", result.Task.ID, result.Error)
	}

	fmt.Printf("Artifacts saved to: .cooperations/handoffs/%s.json\n", result.Task.ID)

	// Write code to output file if specified
	if outputFile != "" && result.Artifacts.Code != "" {
		code := extractCode(result.Artifacts.Code)
		if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		fmt.Printf("Code written to: %s\n", outputFile)
	}

	if verbose && result.Artifacts.Code != "" {
		fmt.Println("\n--- Generated Code ---")
		fmt.Println(result.Artifacts.Code)
	}

	return nil
}

// extractCode extracts code from markdown code blocks if present.
func extractCode(content string) string {
	// Check if content is wrapped in markdown code block
	if len(content) > 6 && content[:3] == "```" {
		// Find the end of the first line (language identifier)
		start := 3
		for start < len(content) && content[start] != '\n' {
			start++
		}
		if start < len(content) {
			start++ // skip the newline
		}

		// Find the closing ```
		end := len(content) - 1
		for end > start && content[end] != '`' {
			end--
		}
		// Move back to before the closing ```
		if end > start+2 && content[end-1] == '`' && content[end-2] == '`' {
			end -= 2
		}
		// Trim trailing newline before ```
		for end > start && (content[end-1] == '\n' || content[end-1] == '\r') {
			end--
		}

		if end > start {
			return content[start:end]
		}
	}
	return content
}

func showStatus(cmd *cobra.Command, args []string) error {
	config := orchestrator.DefaultWorkflowConfig()
	orch, err := orchestrator.New(config)
	if err != nil {
		return fmt.Errorf("initialize orchestrator: %w", err)
	}

	if len(args) == 0 {
		// Show most recent task
		tasks, err := orch.ListTasks()
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			fmt.Println("No tasks found")
			return nil
		}
		task := tasks[len(tasks)-1]
		printTaskInfo(task.ID, task.Status, task.CreatedAt, task.Description)
		return nil
	}

	// Show specific task
	task, err := orch.GetTask(args[0])
	if err != nil {
		return err
	}
	printTaskInfo(task.ID, task.Status, task.CreatedAt, task.Description)

	// Show handoffs
	handoffs, err := orch.GetHandoffs(args[0])
	if err != nil {
		return err
	}
	if len(handoffs) > 0 {
		fmt.Printf("\nHandoffs: %d\n", len(handoffs))
		for i, h := range handoffs {
			fmt.Printf("  %d. %s -> %s (%s, %d tokens)\n", i+1, h.FromRole, h.ToRole, h.Metadata.Model, h.Metadata.TokensUsed)
		}
	}

	return nil
}

func showHistory(cmd *cobra.Command, args []string) error {
	config := orchestrator.DefaultWorkflowConfig()
	orch, err := orchestrator.New(config)
	if err != nil {
		return fmt.Errorf("initialize orchestrator: %w", err)
	}

	tasks, err := orch.ListTasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	limit, _ := cmd.Flags().GetInt("limit")
	start := 0
	if len(tasks) > limit {
		start = len(tasks) - limit
	}

	fmt.Printf("Recent tasks (showing %d of %d):\n\n", min(limit, len(tasks)), len(tasks))
	for i := start; i < len(tasks); i++ {
		t := tasks[i]
		fmt.Printf("  %s  [%s]  %s\n", t.ID, t.Status, truncate(t.Description, 50))
	}

	return nil
}

func runGUI(cmd *cobra.Command, args []string) error {
	task := args[0]
	demo, _ := cmd.Flags().GetBool("demo")

	app := gui.NewApp()
	return app.RunWithDemo(task, demo)
}

func runTUI(cmd *cobra.Command, args []string) error {
	demoMode, _ := cmd.Flags().GetBool("demo")

	// Create workflow stream for communication
	workflowStream := stream.NewWorkflowStream()
	defer workflowStream.Close()

	if demoMode {
		// Run demo mode with simulated events
		go demo.Run(workflowStream)
	} else if len(args) > 0 {
		// Run actual workflow with TUI
		task := args[0]
		go runTUIWorkflow(workflowStream, task)
	}

	// Start TUI
	return tui.Run(workflowStream)
}

func runTUIWorkflow(s *stream.WorkflowStream, task string) {
	// Get max cycles from env
	cycles := 2
	if envCycles := os.Getenv("MAX_REVIEW_CYCLES"); envCycles != "" {
		if c, err := strconv.Atoi(envCycles); err == nil {
			cycles = c
		}
	}

	config := orchestrator.WorkflowConfig{
		MaxReviewCycles: cycles,
	}

	orch, err := orchestrator.New(config)
	if err != nil {
		s.SendError(fmt.Errorf("initialize orchestrator: %w", err))
		return
	}

	// Setup context
	ctx := context.Background()

	s.SendProgress(stream.ProgressUpdate{
		Percent: 0,
		Stage:   "Starting",
		Message: "Running task: " + task,
	})

	// Run the workflow
	result, err := orch.Run(ctx, task)
	if err != nil {
		s.SendError(err)
		return
	}

	if result.Success {
		s.SendToast(stream.ToastNotification{
			Level:   "success",
			Message: "Task completed successfully!",
		})
	} else {
		s.SendToast(stream.ToastNotification{
			Level:   "error",
			Message: "Task failed: " + result.Error,
		})
	}

	s.SignalDone()
}

func printTaskInfo(id, status, createdAt, description string) {
	fmt.Printf("Task: %s\n", id)
	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Created: %s\n", createdAt)
	fmt.Printf("Description: %s\n", description)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
