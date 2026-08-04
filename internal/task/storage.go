package task

//go:generate go run go.uber.org/mock/mockgen@v0.5.0 -source=storage.go -destination=mock_storage_test.go -package=task

type Storage interface {
	AddTask(task *Task) error
	GetTask(uid string) (*Task, error)
	DeleteTask(uid string) error
	CreateAsync(task *Task)
	DeleteAsync(uid string)
	PendingAsyncOperations() int
}
