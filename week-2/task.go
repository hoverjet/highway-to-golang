package main

import (
	"github.com/google/uuid"
	"log/slog"
	"os"
	"time"
)

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

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := NewTaskStore()

	// создаём задачи
	t1 := NewTask("learn go")
	t2 := NewTask("write code")

	// добавляем в store
	store.AddTask(t1)
	store.AddTask(t2)

	// получаем задачу
	if t, ok := store.GetTask(t1.UID); ok {
		logger.Info("before update", "uid", t.UID, "text", t.Text)

		t.SetText("learn go pointers")

		logger.Info("after update", "uid", t.UID, "text", t.Text)
	}

	// удаляем задачу
	store.DeleteTask(t2.UID)

	// проверяем удаление
	if _, ok := store.GetTask(t2.UID); !ok {
		logger.Info("task deleted", "uid", t2.UID)
	}
}
