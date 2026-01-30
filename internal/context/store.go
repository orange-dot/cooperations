package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cooperations/internal/types"
)

// Store manages local JSON file storage for tasks and handoffs.
type Store struct {
	baseDir      string
	generatedDir string
}

// NewStore creates a new store with the given base and generated directories.
func NewStore(baseDir string, generatedDir string) (*Store, error) {
	// Create directories if they don't exist
	handoffsDir := filepath.Join(baseDir, "handoffs")
	if err := os.MkdirAll(handoffsDir, 0755); err != nil {
		return nil, fmt.Errorf("create handoffs directory: %w", err)
	}

	if generatedDir == "" {
		generatedDir = "generated"
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return nil, fmt.Errorf("create generated directory: %w", err)
	}

	return &Store{baseDir: baseDir, generatedDir: generatedDir}, nil
}

// tasksFile returns the path to the tasks.json file.
func (s *Store) tasksFile() string {
	return filepath.Join(s.baseDir, "tasks.json")
}

// handoffFile returns the path to a task's handoff file.
func (s *Store) handoffFile(taskID string) string {
	return filepath.Join(s.baseDir, "handoffs", taskID+".json")
}

// SaveTask saves or updates a task.
func (s *Store) SaveTask(task types.Task) error {
	tasks, err := s.LoadTasks()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Update existing or append new
	found := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			found = true
			break
		}
	}
	if !found {
		tasks = append(tasks, task)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}

	return os.WriteFile(s.tasksFile(), data, 0644)
}

// LoadTasks loads all tasks from storage.
func (s *Store) LoadTasks() ([]types.Task, error) {
	data, err := os.ReadFile(s.tasksFile())
	if err != nil {
		if os.IsNotExist(err) {
			return []types.Task{}, nil
		}
		return nil, fmt.Errorf("read tasks file: %w", err)
	}

	var tasks []types.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}

	return tasks, nil
}

// GetTask retrieves a task by ID.
func (s *Store) GetTask(id string) (*types.Task, error) {
	tasks, err := s.LoadTasks()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == id {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("task not found: %s", id)
}

// SaveHandoff appends a handoff to a task's handoff history.
func (s *Store) SaveHandoff(taskID string, handoff types.Handoff) error {
	handoffs, err := s.LoadHandoffs(taskID)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	handoffs = append(handoffs, handoff)

	data, err := json.MarshalIndent(handoffs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal handoffs: %w", err)
	}

	return os.WriteFile(s.handoffFile(taskID), data, 0644)
}

// LoadHandoffs loads all handoffs for a task.
func (s *Store) LoadHandoffs(taskID string) ([]types.Handoff, error) {
	data, err := os.ReadFile(s.handoffFile(taskID))
	if err != nil {
		if os.IsNotExist(err) {
			return []types.Handoff{}, nil
		}
		return nil, fmt.Errorf("read handoffs file: %w", err)
	}

	var handoffs []types.Handoff
	if err := json.Unmarshal(data, &handoffs); err != nil {
		return nil, fmt.Errorf("unmarshal handoffs: %w", err)
	}

	return handoffs, nil
}

// CreateTask creates a new task with the given description.
func (s *Store) CreateTask(description string) (types.Task, error) {
	task := types.Task{
		ID:          generateID(),
		Description: description,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:      types.TaskStatusPending,
	}

	if err := s.SaveTask(task); err != nil {
		return types.Task{}, err
	}

	return task, nil
}

// UpdateTaskStatus updates a task's status.
func (s *Store) UpdateTaskStatus(taskID string, status string) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = status
	return s.SaveTask(*task)
}

// generateID creates a simple unique ID.
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// TaskOutputDir returns the path to a task's generated output directory.
func (s *Store) TaskOutputDir(taskID string) string {
	return filepath.Join(s.generatedDir, taskID)
}

// EnsureTaskOutputDir creates the task output directory if it doesn't exist.
func (s *Store) EnsureTaskOutputDir(taskID string) (string, error) {
	dir := s.TaskOutputDir(taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create task output directory: %w", err)
	}
	return dir, nil
}

// SaveGeneratedCode saves generated code to the task output directory.
func (s *Store) SaveGeneratedCode(taskID string, filename string, code string) (string, error) {
	dir, err := s.EnsureTaskOutputDir(taskID)
	if err != nil {
		return "", err
	}

	// Create code subdirectory
	codeDir := filepath.Join(dir, "code")
	if err := os.MkdirAll(codeDir, 0755); err != nil {
		return "", fmt.Errorf("create code directory: %w", err)
	}

	// Default filename if not provided
	if filename == "" {
		filename = "main.go"
	}

	path := filepath.Join(codeDir, filepath.Clean(filename))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create code subdirectory: %w", err)
	}
	if err := os.WriteFile(path, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("write code file: %w", err)
	}

	return path, nil
}

// SaveDesignDoc saves a design document to the task output directory.
func (s *Store) SaveDesignDoc(taskID string, content string) (string, error) {
	dir, err := s.EnsureTaskOutputDir(taskID)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "design.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write design doc: %w", err)
	}

	return path, nil
}

// SaveReviewFeedback saves review feedback to the task output directory.
func (s *Store) SaveReviewFeedback(taskID string, content string) (string, error) {
	dir, err := s.EnsureTaskOutputDir(taskID)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "review.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write review feedback: %w", err)
	}

	return path, nil
}

// SaveTaskSummary saves a summary of the task to the output directory.
func (s *Store) SaveTaskSummary(taskID string, task types.Task, artifacts types.HArtifacts) (string, error) {
	dir, err := s.EnsureTaskOutputDir(taskID)
	if err != nil {
		return "", err
	}

	summary := fmt.Sprintf(`# Task Summary

**ID:** %s
**Description:** %s
**Created:** %s
**Status:** %s

## Artifacts

`, task.ID, task.Description, task.CreatedAt, task.Status)

	if artifacts.DesignDoc != "" {
		summary += "- [Design Document](design.md)\n"
	}
	if artifacts.Code != "" {
		summary += "- [Generated Code](code/)\n"
	}
	if artifacts.ReviewFeedback != "" {
		summary += "- [Review Feedback](review.md)\n"
	}
	if artifacts.Notes != "" {
		summary += "\n## Notes\n\n" + artifacts.Notes + "\n"
	}

	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte(summary), 0644); err != nil {
		return "", fmt.Errorf("write task summary: %w", err)
	}

	return path, nil
}
