package memory

import "sync/atomic"

type asyncOperation func()

type asyncExecutor struct {
	pending atomic.Int64
}

func newAsyncExecutor() *asyncExecutor {
	return &asyncExecutor{}
}

func (e *asyncExecutor) enqueue(operation asyncOperation) {
	e.pending.Add(1)

	go func() {
		defer e.pending.Add(-1)

		operation()
	}()
}

func (e *asyncExecutor) PendingOperations() int {
	return int(e.pending.Load())
}
