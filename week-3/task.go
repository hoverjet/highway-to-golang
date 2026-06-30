package main

import (
	"github.com/google/uuid"
	"log/slog"
	"os"
	"time"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -source=task.go -destination=mock_task_storage_test.go -package=main

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
	AddTask(task *Task)
	GetTask(uid string) (*Task, bool)
	DeleteTask(uid string)
}

type TaskStore struct {
	tasks map[string]*Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*Task),
	}
}

func (s *TaskStore) AddTask(task *Task) {
	s.tasks[task.UID] = task
}

func (s *TaskStore) GetTask(uid string) (*Task, bool) {
	task, ok := s.tasks[uid]
	return task, ok
}

func (s *TaskStore) DeleteTask(uid string) {
	delete(s.tasks, uid)
}

type SliceTaskStore struct {
	tasks []*Task
}

func NewSliceTaskStore() *SliceTaskStore {
	return &SliceTaskStore{}
}

func (s *SliceTaskStore) AddTask(task *Task) {
	for i, storedTask := range s.tasks {
		if storedTask.UID == task.UID {
			s.tasks[i] = task
			return
		}
	}

	s.tasks = append(s.tasks, task)
}

func (s *SliceTaskStore) GetTask(uid string) (*Task, bool) {
	for _, task := range s.tasks {
		if task.UID == uid {
			return task, true
		}
	}

	return nil, false
}

func (s *SliceTaskStore) DeleteTask(uid string) {
	for i, task := range s.tasks {
		if task.UID == uid {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return
		}
	}
}

type TaskService struct {
	storage TaskStorage
}

func NewTaskService(storage TaskStorage) *TaskService {
	return &TaskService{
		storage: storage,
	}
}

func (s *TaskService) CreateTask(text string) *Task {
	task := NewTask(text)
	s.storage.AddTask(task)

	return task
}

func (s *TaskService) GetTask(uid string) (*Task, bool) {
	return s.storage.GetTask(uid)
}

func (s *TaskService) UpdateTaskText(uid string, text string) bool {
	task, ok := s.storage.GetTask(uid)
	if !ok {
		return false
	}

	task.SetText(text)
	s.storage.AddTask(task)

	return true
}

func (s *TaskService) DeleteTask(uid string) {
	s.storage.DeleteTask(uid)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewTaskService(NewTaskStore())

	t1 := service.CreateTask("learn go")
	t2 := service.CreateTask("write code")

	if t, ok := service.GetTask(t1.UID); ok {
		logger.Info("before update", "uid", t.UID, "text", t.Text)

		service.UpdateTaskText(t.UID, "learn go interfaces")

		if updatedTask, ok := service.GetTask(t.UID); ok {
			logger.Info("after update", "uid", updatedTask.UID, "text", updatedTask.Text)
		}
	}

	service.DeleteTask(t2.UID)

	if _, ok := service.GetTask(t2.UID); !ok {
		logger.Info("task deleted", "uid", t2.UID)
	}
}
