package app

import (
	"log/slog"
	"time"

	"highway-to-golang/internal/task"
	"highway-to-golang/internal/task/storage/memory"
)

func Run(logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	storage := memory.NewStore()
	service := task.NewService(storage)

	t1, err := service.CreateTask("learn go")
	if err != nil {
		return err
	}

	t2, err := service.CreateTask("write code")
	if err != nil {
		return err
	}

	currentTask, err := service.GetTask(t1.UID)
	if err != nil {
		return err
	}

	logger.Info("before update", "uid", currentTask.UID, "text", currentTask.Text)

	if err := service.UpdateTaskText(currentTask.UID, "learn go interfaces"); err != nil {
		return err
	}

	updatedTask, err := service.GetTask(currentTask.UID)
	if err != nil {
		return err
	}

	logger.Info("after update", "uid", updatedTask.UID, "text", updatedTask.Text)

	if _, err := service.GetTask("missing-task"); err != nil {
		logger.Error("failed to get missing task", "error", err)
	}

	if err := storage.AddTask(t1); err != nil {
		logger.Error("failed to add duplicate task", "error", err)
	}

	if err := service.DeleteTask(t2.UID); err != nil {
		return err
	}

	if _, err := service.GetTask(t2.UID); err != nil && task.IsStorageErrorCode(err, task.ErrKeyNotFoundCode) {
		logger.Info("task deleted", "uid", t2.UID)
	}

	if err := service.DeleteTask("missing-task"); err != nil {
		return err
	}

	logger.Info("missing task delete is idempotent")

	asyncTask := service.CreateAsync("learn goroutines")
	logger.Info("async create queued", "pending_async_operations", service.PendingAsyncOperations())

	time.Sleep(10 * time.Millisecond)

	if storedTask, err := service.GetTask(asyncTask.UID); err == nil {
		logger.Info("async task created", "uid", storedTask.UID, "text", storedTask.Text)
	}

	service.DeleteAsync(asyncTask.UID)
	logger.Info("async delete queued", "pending_async_operations", service.PendingAsyncOperations())

	time.Sleep(10 * time.Millisecond)

	if _, err := service.GetTask(asyncTask.UID); err != nil && task.IsStorageErrorCode(err, task.ErrKeyNotFoundCode) {
		logger.Info("async task deleted", "uid", asyncTask.UID)
	}

	return nil
}
