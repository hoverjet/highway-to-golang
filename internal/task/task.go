package task

import (
	"time"

	"github.com/google/uuid"
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
