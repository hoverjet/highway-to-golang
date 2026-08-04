package task

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestServiceDeleteTaskIsIdempotent(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := NewMockStorage(ctrl)
	service := NewService(storage)

	storage.EXPECT().
		DeleteTask("missing").
		Return(&StorageError{Code: ErrKeyNotFoundCode, Key: "missing"})

	if err := service.DeleteTask("missing"); err != nil {
		t.Fatalf("expected idempotent delete, got %v", err)
	}
}

func TestServiceDeleteTaskReturnsUnexpectedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := NewMockStorage(ctrl)
	service := NewService(storage)
	expectedErr := errors.New("storage is unavailable")

	storage.EXPECT().
		DeleteTask("task-1").
		Return(expectedErr)

	err := service.DeleteTask("task-1")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped storage error, got %v", err)
	}
}
