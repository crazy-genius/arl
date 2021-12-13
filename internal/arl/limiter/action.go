package limiter

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	rc "github.com/go-redis/redis/v8"
)

// CounterNotFound accrued when no counter found in storage
var CounterNotFound = errors.New("counter not found")

// Storage describe storage access contract
type Storage interface {
	// Inc increments storage for provided key => timestamp
	Inc(ctx context.Context, key string, ts int64) error
	// Count receive counter for provided key => timestamp can throw `CounterNotFound`
	Count(ctx context.Context, key string, ts int64) (uint32, error)
	// CountAll receive sum of all counters for provided key can throw `CounterNotFound`
	CountAll(ctx context.Context, key string) (uint32, error)
}

// Memory represents Last Recently Use data and holds staff in memory
type Memory struct {
	lock  sync.RWMutex
	cache map[string]map[int64]uint32
}

func (l *Memory) Inc(_ context.Context, key string, ts int64) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, ok := l.cache[key]

	if !ok {
		l.cache[key] = map[int64]uint32{
			ts: 1,
		}
		return nil
	}

	if _, ok = l.cache[key][ts]; !ok {
		l.cache[key][ts] = 1
	} else {
		l.cache[key][ts] += 1
	}

	return nil
}

func (l *Memory) Count(_ context.Context, key string, ts int64) (uint32, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, ok := l.cache[key]
	if !ok {
		return 0, CounterNotFound
	}

	if cnt, ok := val[ts]; ok {
		return cnt, nil
	}

	return 0, CounterNotFound
}

func (l *Memory) CountAll(_ context.Context, key string) (uint32, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, ok := l.cache[key]
	if !ok {
		return 0, CounterNotFound
	}

	var cnt uint32
	for _, counter := range val {
		cnt += counter
	}

	return cnt, nil
}

// NewInMemoryStorage create new instance of Memory
func NewInMemoryStorage() *Memory {
	return &Memory{
		cache: map[string]map[int64]uint32{},
	}
}

type Redis struct {
	client rc.UniversalClient
}

func (r *Redis) Inc(ctx context.Context, key string, ts int64) error {

	exists := r.client.Exists(ctx, key).Val() == 1

	if err := r.client.HIncrBy(ctx, key, strconv.FormatInt(ts, 10), 1).Err(); err != nil {
		return err
	}

	if !exists {
		if err := r.client.Expire(ctx, key, time.Minute*1).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Redis) Count(ctx context.Context, key string, ts int64) (uint32, error) {
	cmd := r.client.HGet(ctx, key, strconv.FormatInt(ts, 10))

	if err := cmd.Err(); err != nil {
		return 0, err
	}

	val, err := cmd.Uint64()
	if err != nil {
		return 0, err
	}

	return uint32(val), nil
}

func (r *Redis) CountAll(ctx context.Context, key string) (uint32, error) {
	cmd := r.client.HGetAll(ctx, key)

	if err := cmd.Err(); err != nil {
		return 0, err
	}

	var cnt uint32
	for _, val := range cmd.Val() {
		counter, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return 0, err
		}

		cnt += uint32(counter)
	}

	return cnt, nil
}

// NewRedisStorage create new instance of Redis
func NewRedisStorage(client rc.UniversalClient) *Redis {

	return &Redis{
		client: client,
	}
}
