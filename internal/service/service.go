package service

import (
	"fmt"
	"time"

	"github.com/middelmatigheid/top-searches/internal/models"
)

type Storage interface {
	Add(search string, searchTime time.Time)
	GetTop() []models.TopSearch
	GetTopSize() int64
}

type Stoplist interface {
	Add(word string)
	Remove(word string)
	Contains(word string) bool
	GetAll() []string
}

type Limiter interface {
	Allow(user string, timestamp time.Time) bool
}

type Metrics interface {
	AddSearch()
	BlockSearch()
	StoplistHit()
	SetTopSize(size int64)
}

type Service struct {
	Timespan int64
	Storage  Storage
	Stoplist Stoplist
	Limiter  Limiter
	Metrics  Metrics
}

func NewService(timespan int64, storage Storage, stoplist Stoplist, limiter Limiter, metrics Metrics) *Service {
	return &Service{
		Timespan: timespan,
		Storage:  storage,
		Stoplist: stoplist,
		Limiter:  limiter,
		Metrics:  metrics,
	}
}

func (s *Service) AddSearch(search string, user string, timestamp time.Time) error {
	if s.Metrics != nil {
		s.Metrics.AddSearch()
	}
	// Checking if the timestamp is valid
	now := time.Now()
	if timestamp.After(now) {
		return fmt.Errorf("Search \"%s\" was sent from future", search)
	} else if now.Sub(timestamp) > time.Second*time.Duration(s.Timespan) {
		return fmt.Errorf("Search \"%s\" at %v is too old", search, timestamp)
	}
	// Checking if the user spamming
	if s.Limiter.Allow(user, timestamp) {
		s.Storage.Add(search, timestamp)
		return nil
	} else {
		if s.Metrics != nil {
			s.Metrics.BlockSearch()
		}
		return fmt.Errorf("User \"%s\" searches too frequently", user)
	}
}

func (s *Service) GetTopN(n int64) ([]models.TopSearch, error) {
	if s.Metrics != nil {
		s.Metrics.SetTopSize(s.Storage.GetTopSize())
	}
	if n <= 0 {
		return nil, fmt.Errorf("Can't get nonpositive amount %d of top searches", n)
	}
	topSearchesRaw := s.Storage.GetTop()
	topSearches := make([]models.TopSearch, 0, n)
	for _, topSearch := range topSearchesRaw {
		if !s.Stoplist.Contains(topSearch.Search) {
			topSearches = append(topSearches, topSearch)
		} else if s.Metrics != nil {
			s.Metrics.StoplistHit()
		}
		if int64(len(topSearches)) == n {
			break
		}
	}
	return topSearches, nil
}

func (s *Service) AddStoplistWord(word string) {
	s.Stoplist.Add(word)
}

func (s *Service) RemoveStoplistWord(word string) {
	s.Stoplist.Remove(word)
}

func (s *Service) StoplistContains(word string) bool {
	return s.Stoplist.Contains(word)
}

func (s *Service) GetStoplist() []string {
	return s.Stoplist.GetAll()
}
