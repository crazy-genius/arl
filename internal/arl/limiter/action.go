package limiter

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	rc "github.com/go-redis/redis/v8"
)

// ErrCounterNotFound accrued when no counter found in storage
var ErrCounterNotFound = errors.New("counter not found")

// Storage describe storage access contract
type Storage interface {
	// Inc increments storage for provided key => timestamp
	Inc(ctx context.Context, key string, ts int64) error
	// Count receive counter for provided key => timestamp can throw `ErrCounterNotFound`
	Count(ctx context.Context, key string, ts int64) (uint32, error)
	// CountAll receive sum of all counters for provided key can throw `ErrCounterNotFound`
	CountAll(ctx context.Context, key string) (uint32, error)
}

// Memory represents Last Recently Use data and holds staff in memory
type Memory struct {
	lock  sync.RWMutex
	cache map[string]map[int64]uint32
}

func (l *Memory) gc(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-ticker.C:
			now := time.Now().UTC()
			l.lock.Lock()
			var keysToClean []string
			for key, data := range l.cache {
				if len(data) == 0 {
					continue
				}

				for ts := range data {
					if now.After(time.Unix(ts, 0).Add(time.Minute)) {
						keysToClean = append(keysToClean, key)
					}

					break
				}
			}
			for _, key := range keysToClean {
				l.cache[key] = map[int64]uint32{}
			}
			l.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// Inc increments api call count for ts
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
		l.cache[key][ts]++
	}

	return nil
}

// Count returns count of api calls for ts
func (l *Memory) Count(_ context.Context, key string, ts int64) (uint32, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, ok := l.cache[key]
	if !ok {
		return 0, ErrCounterNotFound
	}

	if cnt, ok := val[ts]; ok {
		return cnt, nil
	}

	return 0, ErrCounterNotFound
}

// CountAll returns count of api calls for whole hour
func (l *Memory) CountAll(_ context.Context, key string) (uint32, error) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, ok := l.cache[key]
	if !ok {
		return 0, ErrCounterNotFound
	}

	var cnt uint32
	for _, counter := range val {
		cnt += counter
	}

	return cnt, nil
}

// NewInMemoryStorage create new instance of Memory
func NewInMemoryStorage(ctx context.Context) *Memory {
	ms := &Memory{
		cache: map[string]map[int64]uint32{},
	}

	go ms.gc(ctx)

	return ms
}

// Redis implements redis based realisation of storage
type Redis struct {
	client rc.UniversalClient
}

// Inc increments api call count for ts
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

// Count returns count of api calls for ts
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

// CountAll returns count of api calls for whole hour
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
