package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jarv/newsgoat/internal/logging"
)

// DefaultManager implements the Manager interface
type DefaultManager struct {
	maxWorkers int
	tasks      map[string]*Task
	taskQueue  chan *Task
	handlers   map[TaskType]TaskHandler
	events     chan TaskEvent
	workers    []*worker
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
}

// worker represents a worker that executes tasks
type worker struct {
	id      int
	manager *DefaultManager
	ctx     context.Context
}

// NewManager creates a new task manager
func NewManager(maxWorkers int) Manager {
	return &DefaultManager{
		maxWorkers: maxWorkers,
		tasks:      make(map[string]*Task),
		taskQueue:  make(chan *Task, 100), // Buffered channel for task queue
		handlers:   make(map[TaskType]TaskHandler),
		events:     make(chan TaskEvent, 100), // Buffered channel for events
	}
}

// Start starts the task manager and its workers
func (m *DefaultManager) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("task manager is already running")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true

	// Start workers
	m.workers = make([]*worker, m.maxWorkers)
	for i := 0; i < m.maxWorkers; i++ {
		worker := &worker{
			id:      i,
			manager: m,
			ctx:     m.ctx,
		}
		m.workers[i] = worker
		m.wg.Add(1)
		go worker.start()
	}

	return nil
}

// Stop stops the task manager and all workers
func (m *DefaultManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return fmt.Errorf("task manager is not running")
	}

	m.cancel()
	close(m.taskQueue)

	// Don't wait for workers to finish - they will complete in the background
	// This allows for immediate shutdown when the user quits
	go func() {
		m.wg.Wait()
		close(m.events)
	}()

	m.running = false

	return nil
}

// AddTask adds a task to the queue
func (m *DefaultManager) AddTask(task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.Status = TaskStatusPending

	m.mutex.Lock()
	m.tasks[task.ID] = task
	m.mutex.Unlock()

	select {
	case m.taskQueue <- task:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// GetTask retrieves a task by ID
func (m *DefaultManager) GetTask(id string) (*Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

// ListTasks returns all tasks with optional filtering
func (m *DefaultManager) ListTasks(filter TaskFilter) ([]*Task, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var tasks []*Task
	count := 0

	for _, task := range m.tasks {
		if filter.Type != nil && task.Type != *filter.Type {
			continue
		}
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}

		tasks = append(tasks, task)
		count++

		if filter.Limit > 0 && count >= filter.Limit {
			break
		}
	}

	return tasks, nil
}

// Subscribe returns a channel for task events
func (m *DefaultManager) Subscribe() <-chan TaskEvent {
	return m.events
}

// RegisterHandler registers a task handler
func (m *DefaultManager) RegisterHandler(handler TaskHandler) error {
	// Find all task types this handler can handle
	taskTypes := []TaskType{TaskTypeFeedRefresh} // Add more as needed

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, taskType := range taskTypes {
		if handler.CanHandle(taskType) {
			if _, exists := m.handlers[taskType]; exists {
				return fmt.Errorf("handler for task type %s already exists", taskType)
			}
			m.handlers[taskType] = handler
		}
	}

	return nil
}

// publishEvent publishes a task event
func (m *DefaultManager) publishEvent(event TaskEvent) {
	select {
	case m.events <- event:
	default:
		// Event channel is full, drop the event
		logging.Warn("Event channel full, dropping event", "type", event.Type, "taskID", event.TaskID)
	}
}

// RemoveTask removes a task from the manager
func (m *DefaultManager) RemoveTask(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	// Only allow removing tasks that are not running
	if task.Status == TaskStatusRunning {
		return fmt.Errorf("cannot remove running task: %s", id)
	}

	delete(m.tasks, id)
	logging.Debug("Task removed", "taskID", id)
	return nil
}

// ClearFailedTasks removes all failed tasks
func (m *DefaultManager) ClearFailedTasks() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	count := 0
	for id, task := range m.tasks {
		if task.Status == TaskStatusFailed {
			delete(m.tasks, id)
			count++
		}
	}

	return nil
}

// worker methods

// start starts the worker's main loop
func (w *worker) start() {
	defer w.manager.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case task, ok := <-w.manager.taskQueue:
			if !ok {
				// Channel closed, worker should stop
				return
			}
			w.executeTask(task)
		}
	}
}

// executeTask executes a single task
func (w *worker) executeTask(task *Task) {
	// Update task status
	w.manager.mutex.Lock()
	task.Status = TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	w.manager.mutex.Unlock()

	// Publish started event
	w.manager.publishEvent(TaskEvent{
		Type:      TaskEventStarted,
		TaskID:    task.ID,
		TaskType:  task.Type,
		Status:    TaskStatusRunning,
		Data:      task.Data,
		Timestamp: time.Now(),
	})

	// Find and execute handler
	w.manager.mutex.RLock()
	handler, exists := w.manager.handlers[task.Type]
	w.manager.mutex.RUnlock()

	if !exists {
		w.completeTaskWithError(task, fmt.Errorf("no handler found for task type: %s", task.Type))
		return
	}

	// Execute the task
	err := handler.Execute(w.ctx, task)

	if err != nil {
		w.completeTaskWithError(task, err)
	} else {
		w.completeTask(task)
	}
}

// completeTask marks a task as completed successfully
func (w *worker) completeTask(task *Task) {
	w.manager.mutex.Lock()
	task.Status = TaskStatusCompleted
	now := time.Now()
	task.EndedAt = &now
	w.manager.mutex.Unlock()

	w.manager.publishEvent(TaskEvent{
		Type:      TaskEventCompleted,
		TaskID:    task.ID,
		TaskType:  task.Type,
		Status:    TaskStatusCompleted,
		Data:      task.Data,
		Timestamp: time.Now(),
	})
}

// completeTaskWithError marks a task as failed
func (w *worker) completeTaskWithError(task *Task, err error) {
	w.manager.mutex.Lock()
	task.Status = TaskStatusFailed
	task.Error = err.Error()
	now := time.Now()
	task.EndedAt = &now
	w.manager.mutex.Unlock()

	w.manager.publishEvent(TaskEvent{
		Type:      TaskEventFailed,
		TaskID:    task.ID,
		TaskType:  task.Type,
		Status:    TaskStatusFailed,
		Data:      task.Data,
		Error:     err.Error(),
		Timestamp: time.Now(),
	})

	logging.Error("Task failed", "taskID", task.ID, "type", task.Type, "error", err)
}
