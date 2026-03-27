package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration
type Config struct {
	GithubTokens         []string // Multiple tokens for rotation
	CacheSeconds         int
	Whitelist            []string
	GistWhitelist        []string
	ExcludeRepo          []string
	FetchMultiPageStars  bool
	IconList             []string
	Debug                bool
	AssetsPath           string
	RateLimitPerUser     int // Max requests per user per hour
}

// Load loads configuration from environment variables
func Load() *Config {
	cacheSeconds, _ := strconv.Atoi(getEnv("CACHE_SECONDS", "86400")) // Default 24 hours
	if cacheSeconds < 3600 {
		cacheSeconds = 3600 // Minimum 1 hour
	}
	if cacheSeconds > 86400 {
		cacheSeconds = 86400 // Maximum 24 hours
	}

	fetchMultiPageStars := getEnv("FETCH_MULTI_PAGE_STARS", "false") == "true"
	debug := getEnv("DEBUG", "false") == "true"

	// Parse multiple tokens (comma-separated)
	tokens := parseTokens(getEnv("PAT_1", ""))

	// Rate limit per user per hour (default 30 requests)
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT_PER_USER", "30"))
	if rateLimit < 1 {
		rateLimit = 30
	}

	return &Config{
		GithubTokens:        tokens,
		CacheSeconds:        cacheSeconds,
		Whitelist:           splitEnv("WHITELIST"),
		GistWhitelist:       splitEnv("GIST_WHITELIST"),
		ExcludeRepo:         splitEnv("EXCLUDE_REPO"),
		FetchMultiPageStars: fetchMultiPageStars,
		IconList:            splitEnv("ICON_LIST"),
		Debug:               debug,
		AssetsPath:          getEnv("ASSETS_PATH", "assets/icons"),
		RateLimitPerUser:    rateLimit,
	}
}

// parseTokens parses comma-separated tokens
func parseTokens(tokenStr string) []string {
	if tokenStr == "" {
		return nil
	}
	parts := strings.Split(tokenStr, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	// Split by comma and trim spaces
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
