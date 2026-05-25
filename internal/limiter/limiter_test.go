package limiter

import (
	"testing"
	"time"
)

func TestAllow_FirstRequest(t *testing.T) {
	l, _ := NewLimiter(5)
	now := time.Now()

	allowed := l.Allow("user1", now)

	if !allowed {
		t.Error("First request should be allowed")
	}
}

func TestAllow_FrequentRequests(t *testing.T) {
	l, _ := NewLimiter(5)
	now := time.Now()

	l.Allow("user", now)
	allowed := l.Allow("user", now.Add(1*time.Second))

	if allowed {
		t.Error("Second request within cooldown should be blocked")
	}
}

func TestAllow_RequestAfterCooldown(t *testing.T) {
	l, _ := NewLimiter(5)
	now := time.Now()

	l.Allow("user1", now)
	time.Sleep(6 * time.Second)
	allowed := l.Allow("user1", time.Now())

	if !allowed {
		t.Error("Request after cooldown should be allowed")
	}
}

func TestAllow_DifferentUsers(t *testing.T) {
	l, _ := NewLimiter(5)
	now := time.Now()

	l.Allow("user1", now)
	allowed := l.Allow("user2", now.Add(1*time.Second))

	if !allowed {
		t.Error("Different user should be allowed even if another user is on cooldown")
	}
}
