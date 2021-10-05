package lock

import (
	"context"
	"errors"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
)

var (
	Providers     = map[string]NewLockFunc{}
	ErrLockFailed = errors.New("lock failed")
	ErrLockClosed = errors.New("lock closed")
)

type NewLockFunc func(opts *Options) (Lock, error)

type Options struct {
	Config interface{}
	Logger logger.Logger
}

type Lock interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Lock(ctx context.Context) (err error)
	Unlock(ctx context.Context) (err error)
	HasLock() bool
}

// Acquire attempts to acquire the lock until lock is closed or
// an error that cannot be recovered from occurs
func Acquire(ctx context.Context, lock Lock) error {
	err := lock.Lock(ctx)

	for {
		switch err {
		case ErrLockFailed:
			err = lock.Lock(ctx)
		case ErrLockClosed:
			return nil
		default:
			return err
		}
	}
}
