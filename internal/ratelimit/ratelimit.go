package ratelimit

import (
	"sync"
	"time"
)

// Manager handles rate limiting per user
type Manager struct {
	mu         sync.RWMutex
	userCounts map[string]*userLimit
	limit      int
	window     time.Duration
}

type userLimit struct {
	count     int
	resetTime time.Time
}

// New creates a new rate limit manager
func New(limit int, window time.Duration) *Manager {
	return &Manager{
		userCounts: make(map[string]*userLimit),
		limit:      limit,
		window:     window,
	}
}

// Check checks if a user has exceeded their rate limit
func (m *Manager) Check(userID string) (bool, int, time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	ul, exists := m.userCounts[userID]

	if !exists || now.After(ul.resetTime) {
		// New user or window expired, reset counter
		m.userCounts[userID] = &userLimit{
			count:     1,
			resetTime: now.Add(m.window),
		}
		return true, m.limit - 1, now.Add(m.window)
	}

	if ul.count >= m.limit {
		return false, 0, ul.resetTime
	}

	ul.count++
	return true, m.limit - ul.count, ul.resetTime
}

// GetRemaining returns remaining requests for a user
func (m *Manager) GetRemaining(userID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ul, exists := m.userCounts[userID]
	if !exists {
		return m.limit
	}

	if time.Now().After(ul.resetTime) {
		return m.limit
	}

	return m.limit - ul.count
}
