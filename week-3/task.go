package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -source=task.go -destination=mock_task_storage_test.go -package=main

type StorageErrorCode string

const (
	ErrKeyNotFoundCode StorageErrorCode = "key not found"
	ErrKeyConflictCode StorageErrorCode = "key conflict"
)

type StorageError struct {
	Code StorageErrorCode
	Key  string
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Key)
}

func isStorageErrorCode(err error, code StorageErrorCode) bool {
	var storageErr *StorageError
	if !errors.As(err, &storageErr) {
		return false
	}

	return storageErr.Code == code
}

type asyncOperation func()

type asyncExecutor struct {
	operations chan asyncOperation
	pending    atomic.Int64
}

func newAsyncExecutor() *asyncExecutor {
	executor := &asyncExecutor{
		operations: make(chan asyncOperation),
	}

	go executor.run()

	return executor
}

func (e *asyncExecutor) enqueue(operation asyncOperation) {
	e.pending.Add(1)

	go func() {
		e.operations <- operation
	}()
}

func (e *asyncExecutor) run() {
	for operation := range e.operations {
		operation()
		e.pending.Add(-1)
	}
}

func (e *asyncExecutor) PendingOperations() int {
	return int(e.pending.Load())
}

type Task struct {
	UID       string
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTask(text string) *Task {
	now := time.Now()

	return &Task{
		UID:       uuid.NewString(),
		Text:      text,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Task) SetText(text string) {
	t.Text = text
	t.UpdatedAt = time.Now()
}

type TaskStorage interface {
	AddTask(task *Task) error
	GetTask(uid string) (*Task, error)
	DeleteTask(uid string) error
	CreateAsync(task *Task)
	DeleteAsync(uid string)
	PendingAsyncOperations() int
}

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
	async *asyncExecutor
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*Task),
		async: newAsyncExecutor(),
	}
}

func (s *TaskStore) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[task.UID]; ok {
		return &StorageError{Code: ErrKeyConflictCode, Key: task.UID}
	}

	s.tasks[task.UID] = task
	return nil
}

func (s *TaskStore) GetTask(uid string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[uid]
	if !ok {
		return nil, &StorageError{Code: ErrKeyNotFoundCode, Key: uid}
	}

	return task, nil
}

func (s *TaskStore) DeleteTask(uid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[uid]; !ok {
		return &StorageError{Code: ErrKeyNotFoundCode, Key: uid}
	}

	delete(s.tasks, uid)
	return nil
}

func (s *TaskStore) CreateAsync(task *Task) {
	s.async.enqueue(func() {
		_ = s.AddTask(task)
	})
}

func (s *TaskStore) DeleteAsync(uid string) {
	s.async.enqueue(func() {
		_ = s.DeleteTask(uid)
	})
}

func (s *TaskStore) PendingAsyncOperations() int {
	return s.async.PendingOperations()
}

type SliceTaskStore struct {
	mu    sync.RWMutex
	tasks []*Task
	async *asyncExecutor
}

func NewSliceTaskStore() *SliceTaskStore {
	return &SliceTaskStore{
		async: newAsyncExecutor(),
	}
}

func (s *SliceTaskStore) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, storedTask := range s.tasks {
		if storedTask.UID == task.UID {
			return &StorageError{Code: ErrKeyConflictCode, Key: task.UID}
		}
	}

	s.tasks = append(s.tasks, task)
	return nil
}

func (s *SliceTaskStore) GetTask(uid string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		if task.UID == uid {
			return task, nil
		}
	}

	return nil, &StorageError{Code: ErrKeyNotFoundCode, Key: uid}
}

func (s *SliceTaskStore) DeleteTask(uid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.tasks {
		if task.UID == uid {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}

	return &StorageError{Code: ErrKeyNotFoundCode, Key: uid}
}

func (s *SliceTaskStore) CreateAsync(task *Task) {
	s.async.enqueue(func() {
		_ = s.AddTask(task)
	})
}

func (s *SliceTaskStore) DeleteAsync(uid string) {
	s.async.enqueue(func() {
		_ = s.DeleteTask(uid)
	})
}

func (s *SliceTaskStore) PendingAsyncOperations() int {
	return s.async.PendingOperations()
}

type TaskService struct {
	storage TaskStorage
}

func NewTaskService(storage TaskStorage) *TaskService {
	return &TaskService{
		storage: storage,
	}
}

func (s *TaskService) CreateTask(text string) (*Task, error) {
	task := NewTask(text)
	if err := s.storage.AddTask(task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (s *TaskService) GetTask(uid string) (*Task, error) {
	task, err := s.storage.GetTask(uid)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}

func (s *TaskService) UpdateTaskText(uid string, text string) error {
	task, err := s.storage.GetTask(uid)
	if err != nil {
		return fmt.Errorf("update task text: %w", err)
	}

	task.SetText(text)
	return nil
}

func (s *TaskService) DeleteTask(uid string) error {
	err := s.storage.DeleteTask(uid)
	if err != nil {
		if isStorageErrorCode(err, ErrKeyNotFoundCode) {
			return nil
		}

		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *TaskService) CreateAsync(text string) *Task {
	task := NewTask(text)
	s.storage.CreateAsync(task)

	return task
}

func (s *TaskService) DeleteAsync(uid string) {
	s.storage.DeleteAsync(uid)
}

func (s *TaskService) PendingAsyncOperations() int {
	return s.storage.PendingAsyncOperations()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	storage := NewTaskStore()
	service := NewTaskService(storage)

	t1, err := service.CreateTask("learn go")
	if err != nil {
		logger.Error("failed to create task", "error", err)
		return
	}

	t2, err := service.CreateTask("write code")
	if err != nil {
		logger.Error("failed to create task", "error", err)
		return
	}

	task, err := service.GetTask(t1.UID)
	if err != nil {
		logger.Error("failed to get task", "error", err)
		return
	}

	logger.Info("before update", "uid", task.UID, "text", task.Text)

	if err := service.UpdateTaskText(task.UID, "learn go interfaces"); err != nil {
		logger.Error("failed to update task", "error", err)
		return
	}

	updatedTask, err := service.GetTask(task.UID)
	if err != nil {
		logger.Error("failed to get updated task", "error", err)
		return
	}

	logger.Info("after update", "uid", updatedTask.UID, "text", updatedTask.Text)

	if _, err := service.GetTask("missing-task"); err != nil {
		logger.Error("failed to get missing task", "error", err)
	}

	if err := storage.AddTask(t1); err != nil {
		logger.Error("failed to add duplicate task", "error", err)
	}

	if err := service.DeleteTask(t2.UID); err != nil {
		logger.Error("failed to delete task", "error", err)
		return
	}

	if _, err := service.GetTask(t2.UID); err != nil && isStorageErrorCode(err, ErrKeyNotFoundCode) {
		logger.Info("task deleted", "uid", t2.UID)
	}

	if err := service.DeleteTask("missing-task"); err != nil {
		logger.Error("failed to delete missing task", "error", err)
		return
	}

	logger.Info("missing task delete is idempotent")

	asyncTask := service.CreateAsync("learn goroutines")
	logger.Info("async create queued", "pending_async_operations", service.PendingAsyncOperations())

	time.Sleep(10 * time.Millisecond)

	if task, err := service.GetTask(asyncTask.UID); err == nil {
		logger.Info("async task created", "uid", task.UID, "text", task.Text)
	}

	service.DeleteAsync(asyncTask.UID)
	logger.Info("async delete queued", "pending_async_operations", service.PendingAsyncOperations())

	time.Sleep(10 * time.Millisecond)

	if _, err := service.GetTask(asyncTask.UID); err != nil && isStorageErrorCode(err, ErrKeyNotFoundCode) {
		logger.Info("async task deleted", "uid", asyncTask.UID)
	}
}
