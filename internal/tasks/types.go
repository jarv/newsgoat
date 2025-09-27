package tasks

import (
	"context"
	"time"
)

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeFeedRefresh TaskType = "feed_refresh"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task represents a unit of work that can be executed
type Task struct {
	ID        string                 `json:"id"`
	Type      TaskType               `json:"type"`
	Status    TaskStatus             `json:"status"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	StartedAt *time.Time             `json:"started_at,omitempty"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// TaskHandler defines the interface for executing tasks
type TaskHandler interface {
	Execute(ctx context.Context, task *Task) error
	CanHandle(taskType TaskType) bool
}

// TaskEvent represents events that occur during task execution
type TaskEvent struct {
	Type      TaskEventType          `json:"type"`
	TaskID    string                 `json:"task_id"`
	TaskType  TaskType               `json:"task_type"`
	Status    TaskStatus             `json:"status"`
	Data      map[string]interface{} `json:"data"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// TaskEventType represents the type of task event
type TaskEventType string

const (
	TaskEventStarted   TaskEventType = "task_started"
	TaskEventCompleted TaskEventType = "task_completed"
	TaskEventFailed    TaskEventType = "task_failed"
	TaskEventProgress  TaskEventType = "task_progress"
)

// Manager defines the interface for the task manager
type Manager interface {
	// Start starts the task manager
	Start(ctx context.Context) error

	// Stop stops the task manager
	Stop() error

	// AddTask adds a task to the queue
	AddTask(task *Task) error

	// GetTask retrieves a task by ID
	GetTask(id string) (*Task, error)

	// ListTasks returns all tasks with optional filtering
	ListTasks(filter TaskFilter) ([]*Task, error)

	// Subscribe to task events
	Subscribe() <-chan TaskEvent

	// RegisterHandler registers a task handler
	RegisterHandler(handler TaskHandler) error

	// RemoveTask removes a task from the manager
	RemoveTask(id string) error

	// ClearFailedTasks removes all failed tasks
	ClearFailedTasks() error
}

// TaskFilter represents filtering options for listing tasks
type TaskFilter struct {
	Type   *TaskType   `json:"type,omitempty"`
	Status *TaskStatus `json:"status,omitempty"`
	Limit  int         `json:"limit,omitempty"`
}