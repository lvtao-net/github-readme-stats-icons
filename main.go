package main

import (
	"log"
	"os"
	"time"

	"github-readme-stats/internal/api"
	"github-readme-stats/internal/blacklist"
	"github-readme-stats/internal/cache"
	"github-readme-stats/internal/config"
	"github-readme-stats/internal/github"
	"github-readme-stats/internal/icons"
	"github-readme-stats/internal/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := config.Load()

	// Initialize cache (disabled in debug mode)
	cacheManager := cache.New(cfg)
	if cfg.Debug {
		log.Println("Debug mode: cache disabled")
	} else {
		log.Printf("Cache enabled: %d seconds", cfg.CacheSeconds)
	}

	// Initialize rate limiter (per user per hour)
	rateLimiter := ratelimit.New(cfg.RateLimitPerUser, time.Hour)
	log.Printf("Rate limit: %d requests per user per hour", cfg.RateLimitPerUser)

	// Initialize blacklist manager (24h ban for invalid users, IP check window 10min)
	blacklistManager := blacklist.New()
	log.Println("Blacklist manager initialized: 404 users banned for 24h, IP threshold: 5 different users/10min")

	// Start background cleanup task for expired blacklist entries
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			blacklistManager.CleanExpired()
		}
	}()

	// Initialize GitHub client with multiple tokens
	githubClient := github.NewClient(cfg.GithubTokens)
	log.Printf("GitHub tokens loaded: %d", len(cfg.GithubTokens))

	// Initialize icon manager with config
	iconManager := icons.NewManager(cfg)

	// Generate icons.json if it doesn't exist
	iconsJSONPath := "icons.json"
	if _, err := os.Stat(iconsJSONPath); os.IsNotExist(err) {
		log.Println("icons.json not found, generating...")
		if err := iconManager.GenerateIconsJSON(iconsJSONPath); err != nil {
			log.Printf("Warning: failed to generate icons.json: %v", err)
		} else {
			log.Printf("icons.json generated successfully with %d icons", len(iconManager.GetAllIcons()))
		}
	}

	// Initialize handlers
	handler := api.NewHandler(githubClient, cacheManager, iconManager, rateLimiter, blacklistManager, cfg)

	// Setup router
	r := gin.Default()

	// Static files - serve assets directory
	r.Static("/assets", "./assets")

	// Default homepage - serve index.html directly
	r.GET("/", func(c *gin.Context) {
		c.File("./assets/index.html")
	})

	// API routes
	r.GET("/api", handler.StatsCard)
	r.GET("/api/top-langs", handler.TopLangsCard)
	r.GET("/api/pin", handler.RepoPinCard)
	r.GET("/api/gist", handler.GistPinCard)
	r.GET("/api/wakatime", handler.WakaTimeCard)
	r.GET("/api/icons", handler.SkillIcons)
	r.GET("/api/icons/list", handler.IconList)
	r.GET("/api/icons/meta", handler.IconMetadata)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Blacklist stats (admin endpoint)
	r.GET("/admin/blacklist/stats", func(c *gin.Context) {
		c.JSON(200, blacklistManager.GetStats())
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
