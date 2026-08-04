package memory

import (
	"sync"

	"highway-to-golang/internal/task"
)

type SliceStore struct {
	mu    sync.Mutex
	tasks []*task.Task
	async *asyncExecutor
}

func NewSliceStore() *SliceStore {
	return &SliceStore{
		async: newAsyncExecutor(),
	}
}

func (s *SliceStore) AddTask(newTask *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, storedTask := range s.tasks {
		if storedTask.UID == newTask.UID {
			return &task.StorageError{Code: task.ErrKeyConflictCode, Key: newTask.UID}
		}
	}

	s.tasks = append(s.tasks, newTask)
	return nil
}

func (s *SliceStore) GetTask(uid string) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, storedTask := range s.tasks {
		if storedTask.UID == uid {
			return storedTask, nil
		}
	}

	return nil, &task.StorageError{Code: task.ErrKeyNotFoundCode, Key: uid}
}

func (s *SliceStore) DeleteTask(uid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, storedTask := range s.tasks {
		if storedTask.UID == uid {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return nil
		}
	}

	return &task.StorageError{Code: task.ErrKeyNotFoundCode, Key: uid}
}

func (s *SliceStore) CreateAsync(newTask *task.Task) {
	s.async.enqueue(func() {
		_ = s.AddTask(newTask)
	})
}

func (s *SliceStore) DeleteAsync(uid string) {
	s.async.enqueue(func() {
		_ = s.DeleteTask(uid)
	})
}

func (s *SliceStore) PendingAsyncOperations() int {
	return s.async.PendingOperations()
}
