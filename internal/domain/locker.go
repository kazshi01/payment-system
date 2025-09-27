package domain

import "context"

type Locker interface {
	TryLock(ctx context.Context, key string, ttlSeconds int) (ok bool, token string, err error)
	Unlock(ctx context.Context, key, token string) error
}
