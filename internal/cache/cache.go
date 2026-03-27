package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github-readme-stats/internal/config"
)

// Manager handles caching of responses
type Manager struct {
	cache   *cache.Cache
	cfg     *config.Config
	disabled bool
}

// New creates a new cache manager
func New(cfg *config.Config) *Manager {
	defaultExpiration := cfg.CacheSeconds
	if defaultExpiration < 1 {
		defaultExpiration = 21600
	}
	
	return &Manager{
		cache:    cache.New(time.Duration(defaultExpiration)*time.Second, 10*time.Minute),
		cfg:      cfg,
		disabled: cfg.Debug, // Disable cache in debug mode
	}
}

// Get retrieves a value from cache
func (m *Manager) Get(key string) (interface{}, bool) {
	if m.disabled {
		return nil, false
	}
	return m.cache.Get(key)
}

// Set stores a value in cache
func (m *Manager) Set(key string, value interface{}, duration time.Duration) {
	if m.disabled {
		return
	}
	m.cache.Set(key, value, duration)
}

// Delete removes a value from cache
func (m *Manager) Delete(key string) {
	if m.disabled {
		return
	}
	m.cache.Delete(key)
}

// IsDisabled returns true if cache is disabled
func (m *Manager) IsDisabled() bool {
	return m.disabled
}
