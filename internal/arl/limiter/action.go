package limiter

import (
	"context"
	"errors"
	"sync"
)

// CounterNotFound accrued when no counter found in storage
var CounterNotFound = errors.New("counter not found")

// Storage describe storage access contract
type Storage interface {
	// Inc increments storage for provided key => timestamp
	Inc(ctx context.Context, key string, ts int64) error
	// Count receive counter for provided key => timestamp can throw `CounterNotFound`
	Count(ctx context.Context, key string, ts int64) (uint8, error)
	// CountAll receive sum of all counters for provided key can throw `CounterNotFound`
	CountAll(ctx context.Context, key string) (uint8, error)
}

// Memory represents Last Recently Use data and holds staff in memory
type Memory struct {
	lock  sync.RWMutex
	cache map[string]map[int64]uint8
}

func (l *Memory) Inc(_ context.Context, key string, ts int64) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, ok := l.cache[key]

	if !ok {
		l.cache[key] = map[int64]uint8{
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

func (l *Memory) Count(_ context.Context, key string, ts int64) (uint8, error) {
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

func (l *Memory) CountAll(_ context.Context, key string) (uint8, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, ok := l.cache[key]
	if !ok {
		return 0, CounterNotFound
	}

	var cnt uint8
	for _, counter := range val {
		cnt += counter
	}

	return cnt, nil
}

// NewInMemoryStorage create new instance of Memory
func NewInMemoryStorage() *Memory {
	return &Memory{
		cache: map[string]map[int64]uint8{},
	}
}

type Redis struct {
}

func (r *Redis) Inc(ctx context.Context, key string, ts int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *Redis) Count(ctx context.Context, key string, ts int64) (uint8, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Redis) CountAll(ctx context.Context, key string) (uint8, error) {
	//TODO implement me
	panic("implement me")
}

// NewRedisStorage create new instance of Redis
func NewRedisStorage() *Redis {
	return &Redis{}
}
