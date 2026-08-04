package memory

import (
	"sync"

	"highway-to-golang/internal/task"
)

type Store struct {
	mu    sync.Mutex
	tasks map[string]*task.Task
	async *asyncExecutor
}

func NewStore() *Store {
	return &Store{
		tasks: make(map[string]*task.Task),
		async: newAsyncExecutor(),
	}
}

func (s *Store) AddTask(newTask *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[newTask.UID]; ok {
		return &task.StorageError{Code: task.ErrKeyConflictCode, Key: newTask.UID}
	}

	s.tasks[newTask.UID] = newTask
	return nil
}

func (s *Store) GetTask(uid string) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storedTask, ok := s.tasks[uid]
	if !ok {
		return nil, &task.StorageError{Code: task.ErrKeyNotFoundCode, Key: uid}
	}

	return storedTask, nil
}

func (s *Store) DeleteTask(uid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[uid]; !ok {
		return &task.StorageError{Code: task.ErrKeyNotFoundCode, Key: uid}
	}

	delete(s.tasks, uid)
	return nil
}

func (s *Store) CreateAsync(newTask *task.Task) {
	s.async.enqueue(func() {
		_ = s.AddTask(newTask)
	})
}

func (s *Store) DeleteAsync(uid string) {
	s.async.enqueue(func() {
		_ = s.DeleteTask(uid)
	})
}

func (s *Store) PendingAsyncOperations() int {
	return s.async.PendingOperations()
}
