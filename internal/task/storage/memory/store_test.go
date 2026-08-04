package memory

import (
	"sync"
	"testing"
	"time"

	"highway-to-golang/internal/task"
)

func TestStorageAddTaskReturnsConflict(t *testing.T) {
	testCases := map[string]task.Storage{
		"map storage":   NewStore(),
		"slice storage": NewSliceStore(),
	}

	for name, storage := range testCases {
		t.Run(name, func(t *testing.T) {
			newTask := task.NewTask("learn go")

			if err := storage.AddTask(newTask); err != nil {
				t.Fatalf("expected first add to succeed, got %v", err)
			}

			err := storage.AddTask(newTask)
			if err == nil {
				t.Fatal("expected conflict error")
			}

			if !task.IsStorageErrorCode(err, task.ErrKeyConflictCode) {
				t.Fatalf("expected conflict error code, got %v", err)
			}
		})
	}
}

func TestStorageConcurrentAccess(t *testing.T) {
	testCases := map[string]task.Storage{
		"map storage":   NewStore(),
		"slice storage": NewSliceStore(),
	}

	for name, storage := range testCases {
		t.Run(name, func(t *testing.T) {
			tasks := make([]*task.Task, 100)

			var wg sync.WaitGroup
			for i := range tasks {
				tasks[i] = task.NewTask("learn concurrency")

				wg.Add(1)
				go func(newTask *task.Task) {
					defer wg.Done()

					if err := storage.AddTask(newTask); err != nil {
						t.Errorf("expected add to succeed, got %v", err)
					}
				}(tasks[i])
			}

			wg.Wait()

			for _, storedTask := range tasks {
				if _, err := storage.GetTask(storedTask.UID); err != nil {
					t.Fatalf("expected task %s to exist, got %v", storedTask.UID, err)
				}
			}
		})
	}
}

func TestStoreAsyncOperations(t *testing.T) {
	store := NewStore()
	newTask := task.NewTask("learn async")

	locked := true
	store.mu.Lock()
	defer func() {
		if locked {
			store.mu.Unlock()
		}
	}()

	store.CreateAsync(newTask)

	waitUntil(t, func() bool {
		return store.PendingAsyncOperations() == 1
	})

	store.mu.Unlock()
	locked = false

	waitUntil(t, func() bool {
		_, err := store.GetTask(newTask.UID)
		return err == nil && store.PendingAsyncOperations() == 0
	})

	store.DeleteAsync(newTask.UID)

	waitUntil(t, func() bool {
		_, err := store.GetTask(newTask.UID)
		return task.IsStorageErrorCode(err, task.ErrKeyNotFoundCode)
	})
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition was not met before timeout")
}
