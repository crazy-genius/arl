package limiter

import (
	"context"
	"errors"
	"time"
)

// UnknownSegment accrued when unknown segment provided
var UnknownSegment = errors.New("unknown segment")

// Segment represents a timespan segment
type Segment byte

const (
	Second Segment = iota
	Hour
)

// Service describe rate limiter service contract
type Service interface {
	Inc(ctx context.Context, key string) error
	// Count return count of requests for segment can throw `UnknownSegment`
	Count(ctx context.Context, key string, seg Segment) (uint32, error)
}

type ServiceImpl struct {
	lru       Storage
	permanent Storage
}

func (s *ServiceImpl) Inc(ctx context.Context, key string) error {

	ts := time.Now().UTC().Unix()

	if err := s.lru.Inc(ctx, key, ts); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		if err := s.permanent.Inc(ctx, key, ts); err != nil {
			//TODO: log extra info
		}
	}()

	return nil
}

func (s *ServiceImpl) Count(ctx context.Context, key string, seg Segment) (uint32, error) {
	switch seg {
	case Second:
		ts := time.Now().UTC().Unix()
		cnt, err := s.lru.Count(ctx, key, ts)
		if err != nil {
			return s.permanent.Count(ctx, key, ts)
		}

		return cnt, nil
	case Hour:
		cnt, err := s.lru.CountAll(ctx, key)
		if err != nil {
			return s.permanent.CountAll(ctx, key)
		}

		return cnt, nil
	}

	return 0, UnknownSegment
}

// NewService instantiate new ServiceImpl instance
func NewService(lru Storage, permanent Storage) *ServiceImpl {
	return &ServiceImpl{
		lru:       lru,
		permanent: permanent,
	}
}
