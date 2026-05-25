package storage

import (
	"errors"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/middelmatigheid/top-searches/internal/models"
)

type Storage struct {
	mu *sync.RWMutex
	// Each bucket stores info about specified amount of seconds of the storage timespan
	bucketTimespan int64
	// Stores total number of appearances for each search within the timespan
	total map[string]int64
	// Each bucket stores number of appearances for specific time of the timespan
	buckets []map[string]int64
	// Stores last time, when each query was being searched
	lastSearched map[string]time.Time
	top          []models.TopSearch
	createdAt    time.Time
	logger       *slog.Logger
}

func NewStorage(timespan int64, bucketTimespan int64, logger *slog.Logger) (*Storage, error) {
	if timespan <= 0 {
		return nil, errors.New("Storage timespan should be positive")
	} else if bucketTimespan <= 0 {
		return nil, errors.New("Bucket timespan should be positive")
	}

	mu := &sync.RWMutex{}
	total := make(map[string]int64)
	var buckets []map[string]int64

	// Timespan is being divided evenly into buckets. There is one extra bucket for convenient cleaning
	// Every period of bucketTimespan the next bucket will be emptied, so there are correct amount of buckets used and we have empty bucket for next bucketTimespan
	for range timespan/bucketTimespan + 1 {
		buckets = append(buckets, make(map[string]int64))
	}
	lastSearched := make(map[string]time.Time)
	createdAt := time.Now()

	storage := &Storage{
		mu:             mu,
		bucketTimespan: bucketTimespan,
		buckets:        buckets,
		total:          total,
		lastSearched:   lastSearched,
		createdAt:      createdAt,
		logger:         logger,
	}

	// Cleaning buckets, updating top
	go func(s *Storage) {
		ticker := time.NewTicker(time.Second * time.Duration(bucketTimespan))
		for timestamp := range ticker.C {
			// Preparing next bucket to be used
			s.cleanBucket(timestamp.Add(time.Second * time.Duration(bucketTimespan)))
			s.refreshTop()
		}
	}(storage)

	return storage, nil
}

func (s *Storage) cleanBucket(timestamp time.Time) {
	curSeconds := timestamp.Unix() - s.createdAt.Unix()
	bucketNum := (curSeconds / s.bucketTimespan) % int64(len(s.buckets))

	s.mu.Lock()
	defer s.mu.Unlock()

	for search, count := range s.buckets[bucketNum] {
		s.total[search] -= count
		if s.total[search] <= 0 {
			delete(s.total, search)
		}
	}
	s.buckets[bucketNum] = make(map[string]int64)
	if s.logger != nil {
		s.logger.Debug("Bucket was successfully cleaned", "bucketNum", bucketNum)
	}
}

func (s *Storage) refreshTop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.top = make([]models.TopSearch, 0, len(s.total))
	for search, count := range s.total {
		s.top = append(s.top, models.NewTopSearch(search, count, s.lastSearched[search]))
	}

	slices.SortFunc(s.top, models.CmpTopSearches)
	if s.logger != nil {
		s.logger.Debug("Top was successfully refreshed")
	}
}

func (s *Storage) Add(search string, timestamp time.Time) {
	curSeconds := timestamp.Unix() - s.createdAt.Unix()
	bucketNum := (curSeconds / s.bucketTimespan) % int64(len(s.buckets))

	s.mu.Lock()
	defer s.mu.Unlock()

	s.buckets[bucketNum][search]++
	s.total[search]++
	s.lastSearched[search] = timestamp
	if s.logger != nil {
		s.logger.Debug("Search was successfully added", "search", search, "timestamp", timestamp)
	}
}

func (s *Storage) GetTop() []models.TopSearch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	amount := int64(len(s.top))
	top := make([]models.TopSearch, amount)
	copy(top, s.top)

	return top
}

func (s *Storage) GetTopSize() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.top))
}
