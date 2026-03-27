package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github-readme-stats/internal/blacklist"
	"github-readme-stats/internal/cache"
	"github-readme-stats/internal/cards"
	"github-readme-stats/internal/config"
	"github-readme-stats/internal/github"
	"github-readme-stats/internal/icons"
	"github-readme-stats/internal/ratelimit"
	"github-readme-stats/internal/themes"
	"github-readme-stats/internal/utils"

	"github.com/gin-gonic/gin"
)

// Handler handles API requests
type Handler struct {
	github         *github.Client
	cache          *cache.Manager
	icons          *icons.Manager
	rateLimiter    *ratelimit.Manager
	blacklist      *blacklist.Manager
	config         *config.Config
}

// NewHandler creates a new handler
func NewHandler(github *github.Client, cache *cache.Manager, icons *icons.Manager, rateLimiter *ratelimit.Manager, blacklist *blacklist.Manager, cfg *config.Config) *Handler {
	return &Handler{
		github:      github,
		cache:       cache,
		icons:       icons,
		rateLimiter: rateLimiter,
		blacklist:   blacklist,
		config:      cfg,
	}
}

// checkBlacklist 检查IP和用户是否在黑名单中
func (h *Handler) checkBlacklist(c *gin.Context, username string) bool {
	clientIP := c.ClientIP()
	
	// 检查IP黑名单
	if banned, reason, expireAt := h.blacklist.IsIPBanned(clientIP); banned {
		c.Header("X-Blacklist-Reason", reason)
		c.Header("X-Blacklist-Expire", expireAt.Format(time.RFC3339))
		c.String(http.StatusForbidden, fmt.Sprintf("IP has been blocked. Reason: %s. Expires at: %s", reason, expireAt.Format(time.RFC3339)))
		return false
	}
	
	// 检查用户名黑名单
	if banned, reason, expireAt := h.blacklist.IsUserBanned(username); banned {
		c.Header("X-Blacklist-Reason", reason)
		c.Header("X-Blacklist-Expire", expireAt.Format(time.RFC3339))
		c.String(http.StatusForbidden, fmt.Sprintf("User has been blocked. Reason: %s. Expires at: %s", reason, expireAt.Format(time.RFC3339)))
		return false
	}
	
	// 记录IP-用户名关联，检查是否需要封禁IP
	if shouldBan := h.blacklist.RecordIPUser(clientIP, username); shouldBan {
		h.blacklist.BanIP(clientIP, fmt.Sprintf("Too many different usernames (%d+) from same IP within time window", 5))
		c.String(http.StatusForbidden, "IP has been blocked due to suspicious activity (too many different usernames)")
		return false
	}
	
	return true
}

// banUser 将用户加入黑名单
func (h *Handler) banUser(username, reason string) {
	h.blacklist.BanUser(username, reason)
}

// parseTheme parses theme from request
func (h *Handler) parseTheme(c *gin.Context) themes.Theme {
	themeName := c.DefaultQuery("theme", "default")
	theme := themes.GetTheme(themeName)
	
	// Override with custom colors
	if titleColor := c.Query("title_color"); titleColor != "" {
		theme.TitleColor = themes.ParseColor(titleColor)
	}
	if textColor := c.Query("text_color"); textColor != "" {
		theme.TextColor = themes.ParseColor(textColor)
	}
	if iconColor := c.Query("icon_color"); iconColor != "" {
		theme.IconColor = themes.ParseColor(iconColor)
	}
	if borderColor := c.Query("border_color"); borderColor != "" {
		theme.BorderColor = themes.ParseColor(borderColor)
	}
	if bgColor := c.Query("bg_color"); bgColor != "" {
		theme.BGColor = themes.ParseColor(bgColor)
	}
	if ringColor := c.Query("ring_color"); ringColor != "" {
		theme.RingColor = themes.ParseColor(ringColor)
	}
	
	return theme
}

// StatsCard handles /api endpoint
func (h *Handler) StatsCard(c *gin.Context) {
	// Recover from any panic
	defer func() {
		if r := recover(); r != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %v", r))
		}
	}()

	username := c.Query("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Username is required")
		return
	}

	// Check blacklist (IP and user)
	if !h.checkBlacklist(c, username) {
		return
	}

	// Check rate limit
	allowed, remaining, resetTime := h.rateLimiter.Check(username)
	if !allowed {
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", h.config.RateLimitPerUser))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		c.String(http.StatusTooManyRequests, fmt.Sprintf("Rate limit exceeded. Try again after %s", resetTime.Format(time.RFC3339)))
		return
	}
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", h.config.RateLimitPerUser))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

	// Check whitelist
	if len(h.config.Whitelist) > 0 && !utils.Contains(h.config.Whitelist, username) {
		c.String(http.StatusForbidden, "Username not in whitelist")
		return
	}

	// Cache key
	cacheKey := fmt.Sprintf("stats:%s:%s", username, c.Request.URL.RawQuery)
	
	// Check cache
	if cached, ok := h.cache.Get(cacheKey); ok {
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
		c.String(http.StatusOK, cached.(string))
		return
	}
	
	// Set token from query or use config tokens
	token := c.Query("token")
	if token != "" {
		h.github.SetTokens([]string{token})
	} else {
		h.github.SetTokens(h.config.GithubTokens)
	}

	// Fetch user data
	user, err := h.github.GetUser(username)
	if err != nil {
		// 用户不存在，加入黑名单防止恶意消耗
		h.banUser(username, "User not found (404)")
		c.String(http.StatusNotFound, "User not found")
		return
	}
	
	// Fetch contributions
	contributions, err := h.github.GetContributions(username)
	if err != nil {
		contributions = map[string]int{
			"commits": 0,
			"prs": 0,
			"issues": 0,
		}
	}
	
	// Fetch repos for stars count
	repos, err := h.github.GetAllRepos(username)
	if err == nil {
		totalStars := 0
		for _, repo := range repos {
			totalStars += repo.Stars
		}
		user.TotalStars = totalStars
		user.TotalRepos = len(repos)
	}
	
	// Parse options
	hide := utils.Split(c.Query("hide"))
	show := utils.Split(c.Query("show"))
	
	options := cards.StatsCardOptions{
		Username:          username,
		Hide:              hide,
		Show:              show,
		ShowIcons:         utils.ParseBool(c.DefaultQuery("show_icons", "false")),
		IncludeAllCommits: utils.ParseBool(c.DefaultQuery("include_all_commits", "false")),
		HideRank:          utils.ParseBool(c.DefaultQuery("hide_rank", "false")),
		Theme:             h.parseTheme(c),
		CustomTitle:       c.Query("custom_title"),
		CardWidth:         utils.ParseInt(c.Query("card_width"), 0),
		HideBorder:        utils.ParseBool(c.DefaultQuery("hide_border", "false")),
		BorderRadius:      utils.ParseFloat(c.DefaultQuery("border_radius", "4.5"), 4.5),
		LineHeight:        utils.ParseInt(c.DefaultQuery("line_height", "25"), 25),
		TextBold:          utils.ParseBool(c.DefaultQuery("text_bold", "true")),
		DisableAnimations: utils.ParseBool(c.DefaultQuery("disable_animations", "false")),
		RingColor:         c.DefaultQuery("ring_color", ""),
		NumberFormat:      c.DefaultQuery("number_format", "short"),
		NumberPrecision:   utils.ParseInt(c.Query("number_precision"), -1),
		CommitsYear:       utils.ParseInt(c.Query("commits_year"), 0),
		RankIcon:          c.DefaultQuery("rank_icon", "default"),
	}
	
	// Set default ring color from theme
	if options.RingColor == "" {
		options.RingColor = options.Theme.RingColor
	}
	
	// Render card
	svg := cards.RenderStatsCard(user, contributions, options)
	
	// Cache result
	h.cache.Set(cacheKey, svg, time.Duration(h.config.CacheSeconds)*time.Second)
	
	// Return response
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
	c.String(http.StatusOK, svg)
}

// TopLangsCard handles /api/top-langs endpoint
func (h *Handler) TopLangsCard(c *gin.Context) {
	// Recover from any panic
	defer func() {
		if r := recover(); r != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %v", r))
		}
	}()

	username := c.Query("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Username is required")
		return
	}

	// Check blacklist (IP and user)
	if !h.checkBlacklist(c, username) {
		return
	}

	// Check rate limit
	allowed, remaining, resetTime := h.rateLimiter.Check(username)
	if !allowed {
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", h.config.RateLimitPerUser))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		c.String(http.StatusTooManyRequests, fmt.Sprintf("Rate limit exceeded. Try again after %s", resetTime.Format(time.RFC3339)))
		return
	}
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", h.config.RateLimitPerUser))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

	// Cache key
	cacheKey := fmt.Sprintf("langs:%s:%s", username, c.Request.URL.RawQuery)
	
	// Check cache
	if cached, ok := h.cache.Get(cacheKey); ok {
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
		c.String(http.StatusOK, cached.(string))
		return
	}
	
	// Set token
	token := c.Query("token")
	if token != "" {
		h.github.SetTokens([]string{token})
	} else {
		h.github.SetTokens(h.config.GithubTokens)
	}

	// Parse exclude repos
	excludeRepos := utils.Split(c.Query("exclude_repo"))
	excludeRepos = append(excludeRepos, h.config.ExcludeRepo...)

	// Fetch languages
	languages, err := h.github.GetLanguages(username, excludeRepos)
	if err != nil {
		// 用户不存在或无法访问，加入黑名单
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			h.banUser(username, "User not found when fetching languages (404)")
		}
		errMsg := fmt.Sprintf("Failed to fetch languages: %v", err)
		if strings.Contains(err.Error(), "rate limit") {
			errMsg = "GitHub API rate limit exceeded. Please add a GitHub Token to .env file (PAT_1=your_token)."
		}
		c.String(http.StatusInternalServerError, errMsg)
		return
	}
	
	// Parse options
	hide := utils.Split(c.Query("hide"))
	
	langsCount := utils.ParseInt(c.DefaultQuery("langs_count", "5"), 5)
	langsCount = utils.Clamp(langsCount, 1, 20)
	
	layout := c.DefaultQuery("layout", "normal")
	validLayouts := map[string]bool{
		"normal": true, "compact": true, "donut": true, 
		"donut-vertical": true, "pie": true,
	}
	if !validLayouts[layout] {
		layout = "normal"
	}
	
	options := cards.LangsCardOptions{
		Username:          username,
		Hide:              hide,
		Theme:             h.parseTheme(c),
		CustomTitle:       c.Query("custom_title"),
		CardWidth:         utils.ParseInt(c.Query("card_width"), 0),
		HideBorder:        utils.ParseBool(c.DefaultQuery("hide_border", "false")),
		BorderRadius:      utils.ParseFloat(c.DefaultQuery("border_radius", "4.5"), 4.5),
		Layout:            layout,
		LangsCount:        langsCount,
		HideTitle:         utils.ParseBool(c.DefaultQuery("hide_title", "false")),
		HideProgress:      utils.ParseBool(c.DefaultQuery("hide_progress", "false")),
		StatsFormat:       c.DefaultQuery("stats_format", "percentages"),
		SizeWeight:        utils.ParseFloat(c.DefaultQuery("size_weight", "1"), 1),
		CountWeight:       utils.ParseFloat(c.DefaultQuery("count_weight", "0"), 0),
		DisableAnimations: utils.ParseBool(c.DefaultQuery("disable_animations", "false")),
	}
	
	// Render card
	svg := cards.RenderTopLangsCard(languages, options)
	
	// Cache result
	h.cache.Set(cacheKey, svg, time.Duration(h.config.CacheSeconds)*time.Second)
	
	// Return response
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
	c.String(http.StatusOK, svg)
}

// RepoPinCard handles /api/pin endpoint
func (h *Handler) RepoPinCard(c *gin.Context) {
	username := c.Query("username")
	repoName := c.Query("repo")
	
	if username == "" || repoName == "" {
		c.String(http.StatusBadRequest, "Username and repo are required")
		return
	}
	
	// Cache key
	cacheKey := fmt.Sprintf("repo:%s/%s:%s", username, repoName, c.Request.URL.RawQuery)
	
	// Check cache
	if cached, ok := h.cache.Get(cacheKey); ok {
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
		c.String(http.StatusOK, cached.(string))
		return
	}
	
	// Set token
	token := c.Query("token")
	if token != "" {
		h.github.SetTokens([]string{token})
	} else {
		h.github.SetTokens(h.config.GithubTokens)
	}

	// Fetch repo
	repo, err := h.github.GetRepo(username, repoName)
	if err != nil {
		c.String(http.StatusNotFound, "Repository not found")
		return
	}
	
	// Parse options
	descLines := utils.ParseInt(c.Query("description_lines_count"), 0)
	descLines = utils.Clamp(descLines, 0, 3)
	if descLines == 0 {
		descLines = 2
	}
	
	options := cards.RepoCardOptions{
		Theme:             h.parseTheme(c),
		CustomTitle:       c.Query("custom_title"),
		CardWidth:         utils.ParseInt(c.Query("card_width"), 0),
		HideBorder:        utils.ParseBool(c.DefaultQuery("hide_border", "false")),
		BorderRadius:      utils.ParseFloat(c.DefaultQuery("border_radius", "4.5"), 4.5),
		ShowOwner:         utils.ParseBool(c.DefaultQuery("show_owner", "false")),
		DescriptionLines:  descLines,
		DisableAnimations: utils.ParseBool(c.DefaultQuery("disable_animations", "false")),
	}
	
	// Render card
	svg := cards.RenderRepoCard(repo, options)
	
	// Cache result
	h.cache.Set(cacheKey, svg, time.Duration(h.config.CacheSeconds)*time.Second)
	
	// Return response
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
	c.String(http.StatusOK, svg)
}

// GistPinCard handles /api/gist endpoint
func (h *Handler) GistPinCard(c *gin.Context) {
	gistID := c.Query("id")
	if gistID == "" {
		c.String(http.StatusBadRequest, "Gist ID is required")
		return
	}
	
	// Check gist whitelist
	if len(h.config.GistWhitelist) > 0 && !utils.Contains(h.config.GistWhitelist, gistID) {
		c.String(http.StatusForbidden, "Gist not in whitelist")
		return
	}
	
	// Cache key
	cacheKey := fmt.Sprintf("gist:%s:%s", gistID, c.Request.URL.RawQuery)
	
	// Check cache
	if cached, ok := h.cache.Get(cacheKey); ok {
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
		c.String(http.StatusOK, cached.(string))
		return
	}
	
	// Set token
	token := c.Query("token")
	if token != "" {
		h.github.SetTokens([]string{token})
	} else {
		h.github.SetTokens(h.config.GithubTokens)
	}

	// Fetch gist
	gist, err := h.github.GetGist(gistID)
	if err != nil {
		c.String(http.StatusNotFound, "Gist not found")
		return
	}
	
	// Create a mock repo from gist data for rendering
	repo := &github.Repository{
		Name:        gistID[:8],
		FullName:    gist.Owner.Login + "/" + gistID[:8],
		Description: gist.Description,
		HTMLURL:     gist.HTMLURL,
		Stars:       0,
		Forks:       0,
		Owner: struct {
			Login string `json:"login"`
		}{Login: gist.Owner.Login},
	}
	
	// Get first file's language
	for _, file := range gist.Files {
		repo.Language = file.Language
		break
	}
	
	// Parse options
	options := cards.RepoCardOptions{
		Theme:             h.parseTheme(c),
		CustomTitle:       c.Query("custom_title"),
		CardWidth:         utils.ParseInt(c.Query("card_width"), 0),
		HideBorder:        utils.ParseBool(c.DefaultQuery("hide_border", "false")),
		BorderRadius:      utils.ParseFloat(c.DefaultQuery("border_radius", "4.5"), 4.5),
		ShowOwner:         utils.ParseBool(c.DefaultQuery("show_owner", "false")),
		DescriptionLines:  2,
		DisableAnimations: utils.ParseBool(c.DefaultQuery("disable_animations", "false")),
	}
	
	// Render card
	svg := cards.RenderRepoCard(repo, options)
	
	// Cache result
	h.cache.Set(cacheKey, svg, time.Duration(h.config.CacheSeconds)*time.Second)
	
	// Return response
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
	c.String(http.StatusOK, svg)
}

// WakaTimeCard handles /api/wakatime endpoint
func (h *Handler) WakaTimeCard(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.String(http.StatusBadRequest, "Username is required")
		return
	}
	
	// Cache key
	cacheKey := fmt.Sprintf("wakatime:%s:%s", username, c.Request.URL.RawQuery)
	
	// Check cache
	if cached, ok := h.cache.Get(cacheKey); ok {
		c.Header("Content-Type", "image/svg+xml")
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
		c.String(http.StatusOK, cached.(string))
		return
	}
	
	// Parse options
	layout := c.DefaultQuery("layout", "default")
	if layout != "default" && layout != "compact" {
		layout = "default"
	}
	
	langsCount := utils.ParseInt(c.DefaultQuery("langs_count", "0"), 0)
	
	options := cards.WakaTimeCardOptions{
		Theme:             h.parseTheme(c),
		CustomTitle:       c.Query("custom_title"),
		CardWidth:         utils.ParseInt(c.Query("card_width"), 0),
		HideBorder:        utils.ParseBool(c.DefaultQuery("hide_border", "false")),
		BorderRadius:      utils.ParseFloat(c.DefaultQuery("border_radius", "4.5"), 4.5),
		HideTitle:         utils.ParseBool(c.DefaultQuery("hide_title", "false")),
		HideProgress:      utils.ParseBool(c.DefaultQuery("hide_progress", "false")),
		Layout:            layout,
		LangsCount:        langsCount,
		LineHeight:        utils.ParseInt(c.DefaultQuery("line_height", "25"), 25),
		DisableAnimations: utils.ParseBool(c.DefaultQuery("disable_animations", "false")),
	}
	
	// Mock WakaTime stats (implement actual WakaTime API integration)
	stats := map[string]int{
		"total_time": 0,
		"languages":  0,
	}
	
	// Render card
	svg := cards.RenderWakaTimeCard(stats, options)
	
	// Cache result
	h.cache.Set(cacheKey, svg, time.Duration(h.config.CacheSeconds)*time.Second)
	
	// Return response
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", h.config.CacheSeconds))
	c.String(http.StatusOK, svg)
}

// SkillIcons handles /api/icons endpoint
func (h *Handler) SkillIcons(c *gin.Context) {
	// Get icon list
	iconsParam := c.Query("i")
	if iconsParam == "" {
		c.String(http.StatusBadRequest, "Icons parameter 'i' is required")
		return
	}
	
	var iconNames []string
	
	// Special case: i=all returns all icons
	if strings.ToLower(iconsParam) == "all" {
		iconNames = h.icons.GetAllIcons()
	} else {
		// Parse icons
		iconNames = strings.Split(iconsParam, ",")
	}
	
	// Validate icons
	var validIcons []string
	for _, name := range iconNames {
		name = strings.TrimSpace(strings.ToLower(name))
		if h.icons.IsValidIcon(name) {
			validIcons = append(validIcons, name)
		}
	}
	
	if len(validIcons) == 0 {
		c.String(http.StatusBadRequest, "No valid icons specified")
		return
	}
	
	// Parse theme
	theme := c.DefaultQuery("theme", "dark")
	if theme != "dark" && theme != "light" {
		theme = "dark"
	}
	
	// Parse perline
	perLine := utils.ParseInt(c.DefaultQuery("perline", "15"), 15)
	perLine = utils.Clamp(perLine, 1, 50)
	
	// Generate SVG
	svg := h.icons.GenerateSVG(validIcons, theme, perLine)
	
	// Return response
	c.Header("Content-Type", "image/svg+xml")
	c.Header("Cache-Control", "max-age=86400") // Cache for 24 hours
	c.String(http.StatusOK, svg)
}

// IconList handles /api/icons/list endpoint - returns list of all available icons
func (h *Handler) IconList(c *gin.Context) {
	iconList := h.icons.GetIconsList()
	c.JSON(http.StatusOK, gin.H{
		"icons": iconList,
		"total": len(iconList),
	})
}

// IconMetadata handles /api/icons/meta endpoint - returns detailed icon metadata
func (h *Handler) IconMetadata(c *gin.Context) {
	metadata := h.icons.GetIconsMetadata()
	c.JSON(http.StatusOK, metadata)
}
