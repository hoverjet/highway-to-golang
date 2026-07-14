package task

import (
	"errors"
	"fmt"
)

type StorageErrorCode string

const (
	ErrKeyNotFoundCode StorageErrorCode = "key not found"
	ErrKeyConflictCode StorageErrorCode = "key conflict"
)

type StorageError struct {
	Code StorageErrorCode
	Key  string
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Key)
}

func IsStorageErrorCode(err error, code StorageErrorCode) bool {
	var storageErr *StorageError
	if !errors.As(err, &storageErr) {
		return false
	}

	return storageErr.Code == code
}
