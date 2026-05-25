package storage

import (
	"fmt"
	"testing"
	"time"
)

func TestAdd_SingleSearch(t *testing.T) {
	s, err := NewStorage(300, 10, nil)
	if err != nil {
		fmt.Println(err)
	}
	now := time.Now()

	s.Add("phone", now)
	// Waiting for refreshing top
	time.Sleep(10 * time.Second)

	top := s.GetTop()
	if len(top) != 1 {
		t.Errorf("Expected 1 item, got %d", len(top))
	} else if top[0].Search != "phone" {
		t.Errorf("Expected 'phone', got %s", top[0].Search)
	} else if top[0].Count != 1 {
		t.Errorf("Expected count 1, got %d", top[0].Count)
	}
}

func TestAdd_MultipleSameSearch(t *testing.T) {
	s, _ := NewStorage(300, 10, nil)
	now := time.Now()

	s.Add("phone", now)
	s.Add("phone", now.Add(1*time.Second))

	// Waiting for refreshing top
	time.Sleep(10 * time.Second)

	top := s.GetTop()
	if len(top) != 1 {
		t.Errorf("Expected 1 item, got %d", len(top))
	} else if top[0].Count != 2 {
		t.Errorf("Expected count 2, got %d", top[0].Count)
	}
}

func TestAdd_DifferentSearches(t *testing.T) {
	s, _ := NewStorage(300, 10, nil)
	now := time.Now()

	s.Add("phone", now)
	s.Add("pants", now.Add(1*time.Second))
	s.Add("phone", now.Add(2*time.Second))

	// Waiting for refreshing top
	time.Sleep(10 * time.Second)

	top := s.GetTop()

	if len(top) != 2 {
		t.Errorf("Expected 2 item, got %d", len(top))
	} else if top[0].Search != "phone" {
		t.Errorf("Expected 'phone' first, got %s", top[0].Search)
	} else if top[0].Count != 2 {
		t.Errorf("Expected phone count 2, got %d", top[0].Count)
	}
}
