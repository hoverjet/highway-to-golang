package task

import "fmt"

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) CreateTask(text string) (*Task, error) {
	task := NewTask(text)
	if err := s.storage.AddTask(task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (s *Service) GetTask(uid string) (*Task, error) {
	task, err := s.storage.GetTask(uid)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}

func (s *Service) UpdateTaskText(uid string, text string) error {
	task, err := s.storage.GetTask(uid)
	if err != nil {
		return fmt.Errorf("update task text: %w", err)
	}

	task.SetText(text)
	return nil
}

func (s *Service) DeleteTask(uid string) error {
	err := s.storage.DeleteTask(uid)
	if err != nil {
		if IsStorageErrorCode(err, ErrKeyNotFoundCode) {
			return nil
		}

		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}

func (s *Service) CreateAsync(text string) *Task {
	task := NewTask(text)
	s.storage.CreateAsync(task)

	return task
}

func (s *Service) DeleteAsync(uid string) {
	s.storage.DeleteAsync(uid)
}

func (s *Service) PendingAsyncOperations() int {
	return s.storage.PendingAsyncOperations()
}
