package service

import (
	"testing"
	"time"

	"github.com/middelmatigheid/top-searches/internal/models"
)

type mockStorage struct {
	searches []string
}

func (m *mockStorage) Add(search string, searchTime time.Time) {
	m.searches = append(m.searches, search)
}
func (m *mockStorage) GetTop() []models.TopSearch {
	return []models.TopSearch{
		{Search: "phone", Count: 100},
		{Search: "pants", Count: 80},
	}
}
func (m *mockStorage) GetTopSize() int64 {
	return int64(len(m.searches))
}

type mockStoplist struct {
	words map[string]struct{}
}

func (m *mockStoplist) Contains(word string) bool {
	_, ok := m.words[word]
	return ok
}
func (m *mockStoplist) Add(word string)    {}
func (m *mockStoplist) Remove(word string) {}
func (m *mockStoplist) GetAll() []string   { return nil }

type mockLimiter struct {
	allow bool
}

func (m *mockLimiter) Allow(user string, timestamp time.Time) bool {
	return m.allow
}

func TestAddSearch_Success(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: make(map[string]struct{})}
	limiter := &mockLimiter{allow: true}

	s := NewService(300, storage, stoplist, limiter, nil)

	now := time.Now()

	err := s.AddSearch("phone", "alice", now)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(storage.searches) != 1 {
		t.Errorf("Expected 1 search in storage, got %d", len(storage.searches))
	}
	if storage.searches[0] != "phone" {
		t.Errorf("Expected 'phone', got %s", storage.searches[0])
	}
}

func TestAddSearch_FutureTimestamp(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: make(map[string]struct{})}
	limiter := &mockLimiter{allow: true}

	s := NewService(300, storage, stoplist, limiter, nil)

	future := time.Now().Add(1 * time.Hour)

	err := s.AddSearch("phone", "alice", future)

	if err == nil {
		t.Error("Expected error for future timestamp, got nil")
	}
}

func TestAddSearch_OldTimestamp(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: make(map[string]struct{})}
	limiter := &mockLimiter{allow: true}

	s := NewService(300, storage, stoplist, limiter, nil)

	old := time.Now().Add(-10 * time.Minute)

	err := s.AddSearch("phone", "alice", old)

	if err == nil {
		t.Error("Expected error for too old timestamp, got nil")
	}
}

func TestAddSearch_LimiterHit(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: make(map[string]struct{})}
	limiter := &mockLimiter{allow: false}

	s := NewService(300, storage, stoplist, limiter, nil)

	err := s.AddSearch("phone", "spammer", time.Now())

	if err == nil {
		t.Error("Expected error for hitting the limiter, got nil")
	} else if len(storage.searches) != 0 {
		t.Error("Spam request should not be added to storage")
	}
}

func TestGetTopN(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: make(map[string]struct{})}
	limiter := &mockLimiter{allow: true}

	s := NewService(300, storage, stoplist, limiter, nil)

	top, err := s.GetTopN(1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if len(top) != 1 {
		t.Errorf("Expected 1 item, got %d", len(top))
	} else if top[0].Search != "phone" {
		t.Errorf("Expected 'phone', got %s", top[0].Search)
	}
}

func TestGetTopN_InvalidLimit(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: make(map[string]struct{})}
	limiter := &mockLimiter{allow: true}

	s := NewService(300, storage, stoplist, limiter, nil)

	_, err := s.GetTopN(-5)

	if err == nil {
		t.Error("Expected error for negative limit, got nil")
	}
}

func TestGetTopN_StoplistWord(t *testing.T) {
	storage := &mockStorage{}
	stoplist := &mockStoplist{words: map[string]struct{}{"phone": {}}}
	limiter := &mockLimiter{allow: true}

	s := NewService(300, storage, stoplist, limiter, nil)

	top, err := s.GetTopN(5)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if len(top) != 1 {
		t.Errorf("Expected 1 item, got %d", len(top))
	} else if top[0].Search != "pants" {
		t.Errorf("Expected 'pants', got %s", top[0].Search)
	}
}
