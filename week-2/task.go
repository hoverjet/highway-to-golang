package main

import (
	"fmt"
	"time"
)

type Task struct {
	UID       string
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTask(uid string, text string) *Task {
	now := time.Now()

	return &Task{
		UID:       uid,
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
	store := NewTaskStore()

	// создаём задачи
	t1 := NewTask("1", "learn go")
	t2 := NewTask("2", "write code")

	// добавляем в store
	store.AddTask(t1)
	store.AddTask(t2)

	// получаем задачу
	if t, ok := store.GetTask("1"); ok {
		fmt.Println("before:", t.Text)

		t.SetText("learn go pointers")

		fmt.Println("after:", t.Text)
	}

	// удаляем задачу
	store.DeleteTask("2")

	// проверяем удаление
	if _, ok := store.GetTask("2"); !ok {
		fmt.Println("task 2 deleted")
	}
}
