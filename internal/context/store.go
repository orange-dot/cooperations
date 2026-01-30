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
	baseDir string
}

// NewStore creates a new store with the given base directory.
func NewStore(baseDir string) (*Store, error) {
	// Create directories if they don't exist
	handoffsDir := filepath.Join(baseDir, "handoffs")
	if err := os.MkdirAll(handoffsDir, 0755); err != nil {
		return nil, fmt.Errorf("create handoffs directory: %w", err)
	}

	return &Store{baseDir: baseDir}, nil
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
