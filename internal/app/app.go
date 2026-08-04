package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"highway-to-golang/internal/task"
	"highway-to-golang/internal/task/storage/memory"
)

func Run(ctx context.Context) error {
	storage := memory.NewStore()
	service := task.NewService(storage)

	t1, err := service.CreateTask("learn go")
	if err != nil {
		return fmt.Errorf("create first task: %w", err)
	}

	t2, err := service.CreateTask("write code")
	if err != nil {
		return fmt.Errorf("create second task: %w", err)
	}

	currentTask, err := service.GetTask(t1.UID)
	if err != nil {
		return fmt.Errorf("get task before update: %w", err)
	}

	slog.Info("before update", "uid", currentTask.UID, "text", currentTask.Text)

	if err := service.UpdateTaskText(currentTask.UID, "learn go interfaces"); err != nil {
		return fmt.Errorf("update task text: %w", err)
	}

	updatedTask, err := service.GetTask(currentTask.UID)
	if err != nil {
		return fmt.Errorf("get task after update: %w", err)
	}

	slog.Info("after update", "uid", updatedTask.UID, "text", updatedTask.Text)

	if _, err := service.GetTask("missing-task"); err != nil {
		slog.Error("failed to get missing task", "error", err)
	}

	if err := storage.AddTask(t1); err != nil {
		slog.Error("failed to add duplicate task", "error", err)
	}

	if err := service.DeleteTask(t2.UID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if _, err := service.GetTask(t2.UID); err != nil && task.IsStorageErrorCode(err, task.ErrKeyNotFoundCode) {
		slog.Info("task deleted", "uid", t2.UID)
	}

	if err := service.DeleteTask("missing-task"); err != nil {
		return fmt.Errorf("delete missing task: %w", err)
	}

	slog.Info("missing task delete is idempotent")

	asyncTask := service.CreateAsync("learn goroutines")
	slog.Info("async create queued", "pending_async_operations", service.PendingAsyncOperations())

	if err := wait(ctx, 10*time.Millisecond); err != nil {
		return fmt.Errorf("wait for async task creation: %w", err)
	}

	if storedTask, err := service.GetTask(asyncTask.UID); err == nil {
		slog.Info("async task created", "uid", storedTask.UID, "text", storedTask.Text)
	}

	service.DeleteAsync(asyncTask.UID)
	slog.Info("async delete queued", "pending_async_operations", service.PendingAsyncOperations())

	if err := wait(ctx, 10*time.Millisecond); err != nil {
		return fmt.Errorf("wait for async task deletion: %w", err)
	}

	if _, err := service.GetTask(asyncTask.UID); err != nil && task.IsStorageErrorCode(err, task.ErrKeyNotFoundCode) {
		slog.Info("async task deleted", "uid", asyncTask.UID)
	}

	return nil
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
