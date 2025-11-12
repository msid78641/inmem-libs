package inmem_cache

import (
	"errors"
	"fmt"
)

var (
	ErrEntryNotFound     = errors.New("entry not found")
	ErrStaleResponse     = errors.New("stale response")
	ErrCacheAdaptorNil   = errors.New("cache adaptor is nil")
	ErrLoaderNil         = errors.New("loader function is nil")
	ErrInvalidCacheEntry = errors.New("invalid cache entry")
	ErrLoaderFailed      = errors.New("loader function is nil")

	ErrInvalidDeletionArgs = errors.New("missing deletion keys or tags")
)

func WrapError(wrapper string, err error) error {
	// log it and return the linked or formatted error
	if len(wrapper) > 0 {
		return fmt.Errorf(" %s : %w ", wrapper, err)
	}
	return err
}

const (
	GET    CacheOperation = "get"
	SET    CacheOperation = "set"
	DELETE CacheOperation = "delete"

	SOFTDELETE CacheOperation = "softDelete"
)

type CacheOperation string
type CacheError struct {
	Operation CacheOperation
	Key       string
	BaseError error
}

func (c *CacheError) Error() string {
	return fmt.Sprintf("cache %s failed for key=%q: %v", c.Operation, c.Key, c.BaseError)
}

func cacheError(operation CacheOperation, key string, baseError error) error {
	return &CacheError{
		Operation: operation,
		Key:       key,
		BaseError: baseError,
	}
}
