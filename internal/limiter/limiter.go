package limiter

import (
	"errors"
	"sync"
	"time"
)

type Limiter struct {
	mu       *sync.Mutex
	cooldown int64
	// Stores 2 buckets, every bucket stores users, that searched something within cooldown timespan
	// Every period of cooldown the bucket switches and empties
	buckets   []map[string]struct{}
	createdAt time.Time
}

func NewLimiter(cooldown int64) (*Limiter, error) {
	if cooldown <= 0 {
		return nil, errors.New("Cooldown should be positive")
	}

	mu := &sync.Mutex{}
	buckets := make([]map[string]struct{}, 0, 2)
	for range 2 {
		buckets = append(buckets, make(map[string]struct{}))
	}
	createdAt := time.Now()

	limiter := &Limiter{
		mu:        mu,
		cooldown:  cooldown,
		buckets:   buckets,
		createdAt: createdAt,
	}

	// Cleaning other bucket
	go func(l *Limiter) {
		ticker := time.NewTicker(time.Second * time.Duration(l.cooldown))
		for timestamp := range ticker.C {
			l.clean(timestamp.Add(time.Second * time.Duration(l.cooldown)))
		}
	}(limiter)

	return limiter, nil
}

func (l *Limiter) clean(timestamp time.Time) {
	curSeconds := timestamp.Unix() - l.createdAt.Unix()
	bucketNum := (curSeconds / l.cooldown) % int64(len(l.buckets))

	l.mu.Lock()
	defer l.mu.Unlock()

	l.buckets[bucketNum] = make(map[string]struct{})
}

func (l *Limiter) Allow(user string, timestamp time.Time) bool {
	curSeconds := timestamp.Unix() - l.createdAt.Unix()
	bucketNum := (curSeconds / l.cooldown) % int64(len(l.buckets))

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.buckets[bucketNum][user]; !ok {
		l.buckets[bucketNum][user] = struct{}{}
		return true
	} else {
		return false
	}
}
