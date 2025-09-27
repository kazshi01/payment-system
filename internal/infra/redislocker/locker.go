package redislocker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

type Locker struct{ cli *redis.Client }

func New(addr, password string, db int) *Locker {
	return &Locker{cli: redis.NewClient(&redis.Options{
		Addr: addr, Password: password, DB: db,
	})}
}

func randToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (l *Locker) TryLock(ctx context.Context, key string, ttlSeconds int) (bool, string, error) {
	token, err := randToken()
	if err != nil {
		return false, "", err
	}
	ok, err := l.cli.SetNX(ctx, key, token, time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		return false, "", err
	}
	if !ok {
		return false, "", nil
	} // 既にロックあり
	return true, token, nil
}

// value一致時のみDELするLua
var luaUnlock = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end
`)

func (l *Locker) Unlock(ctx context.Context, key, token string) error {
	_, err := luaUnlock.Run(ctx, l.cli, []string{key}, token).Result()
	return err
}
