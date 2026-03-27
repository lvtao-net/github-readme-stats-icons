package cards

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github-readme-stats/internal/github"
	"github-readme-stats/internal/themes"
	"github-readme-stats/internal/utils"
)

// RenderStatsCard renders a GitHub stats card
type StatsCardOptions struct {
	Username          string
	Hide              []string
	Show              []string
	ShowIcons         bool
	IncludeAllCommits bool
	HideRank          bool
	Theme             themes.Theme
	CustomTitle       string
	CardWidth         int
	HideBorder        bool
	BorderRadius      float64
	LineHeight        int
	TextBold          bool
	DisableAnimations bool
	RingColor         string
	NumberFormat      string
	NumberPrecision   int
	CommitsYear       int
	RankIcon          string
}

// StatsData holds the stats to display
type StatsData struct {
	Name                string
	Value               int
	Icon                string
	Show                bool
}

// RenderStatsCard renders the stats SVG card
func RenderStatsCard(stats *github.UserStats, contributions map[string]int, options StatsCardOptions) string {
	// Prepare data
	data := []StatsData{
		{Name: "Total Stars Earned", Value: stats.TotalStars, Icon: "★", Show: !utils.Contains(options.Hide, "stars")},
		{Name: "Total Commits", Value: contributions["commits"], Icon: "📝", Show: !utils.Contains(options.Hide, "commits")},
		{Name: "Total PRs", Value: contributions["prs"], Icon: "🔀", Show: !utils.Contains(options.Hide, "prs")},
		{Name: "Total Issues", Value: contributions["issues"], Icon: "●", Show: !utils.Contains(options.Hide, "issues")},
		{Name: "Contributed to", Value: stats.TotalContributions, Icon: "📂", Show: !utils.Contains(options.Hide, "contribs")},
	}
	
	// Add optional stats
	if utils.Contains(options.Show, "reviews") {
		data = append(data, StatsData{Name: "Total Reviews", Value: contributions["reviews"], Icon: "👀", Show: true})
	}
	if utils.Contains(options.Show, "discussions_started") {
		data = append(data, StatsData{Name: "Discussions Started", Value: contributions["discussions_started"], Icon: "💬", Show: true})
	}
	if utils.Contains(options.Show, "discussions_answered") {
		data = append(data, StatsData{Name: "Discussions Answered", Value: contributions["discussions_answered"], Icon: "✓", Show: true})
	}
	if utils.Contains(options.Show, "prs_merged") {
		data = append(data, StatsData{Name: "PRs Merged", Value: contributions["prs_merged"], Icon: "🔀", Show: true})
	}
	if utils.Contains(options.Show, "prs_merged_percentage") {
		data = append(data, StatsData{Name: "Merged PRs %", Value: int(stats.PRsMergedPercentage), Icon: "%", Show: true})
	}
	
	// Calculate dimensions
	cardWidth := options.CardWidth
	if cardWidth == 0 {
		cardWidth = 495
	}
	
	// Calculate height based on visible items
	visibleCount := 0
	for _, d := range data {
		if d.Show {
			visibleCount++
		}
	}
	// Calculate height: title(55) + items + padding(30)
	height := 85 + visibleCount*options.LineHeight
	if !options.HideRank {
		// When showing rank circle, need more height
		height = 200
	}
	// Ensure minimum height
	if height < 120 {
		height = 120
	}
	
	title := options.CustomTitle
	if title == "" {
		title = fmt.Sprintf("%s's GitHub Stats", stats.Login)
	}
	
	// Calculate rank
	rank, percentile := utils.CalculateRank(
		contributions["commits"],
		contributions["prs"],
		contributions["issues"],
		contributions["reviews"],
		stats.TotalStars,
		stats.Followers,
	)
	
	// Build SVG
	var sb strings.Builder
	
	// Gradient definition if needed
	bgColor := options.Theme.BGColor
	if strings.Contains(bgColor, ",") {
		// Parse gradient: angle,start,end
		parts := strings.Split(bgColor, ",")
		if len(parts) >= 3 {
			sb.WriteString(fmt.Sprintf(`<defs>
				<linearGradient id="grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
					<stop offset="0%%" style="stop-color:#%s" />
					<stop offset="100%%" style="stop-color:#%s" />
				</linearGradient>
			</defs>`, parts[1], parts[2]))
			bgColor = "url(#grad)"
		}
	} else {
		bgColor = "#" + bgColor
	}
	
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, 
		cardWidth, height, cardWidth, height))
	
	// Background
	borderProps := ""
	if !options.HideBorder {
		borderProps = fmt.Sprintf(` stroke="#%s" stroke-width="1"`, options.Theme.BorderColor)
	}
	sb.WriteString(fmt.Sprintf(`<rect x="0.5" y="0.5" width="%d" height="%d" rx="%.1f" fill="%s"%s/>`,
		cardWidth-1, height-1, options.BorderRadius, bgColor, borderProps))
	
	// Title
	titleY := 35
	sb.WriteString(fmt.Sprintf(`<text x="25" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="18" font-weight="bold" fill="#%s">%s</text>`,
		titleY, options.Theme.TitleColor, utils.EscapeXML(title)))
	
	// Stats
	itemY := 65
	itemX := 25
	iconX := itemX
	textX := itemX + 30
	if !options.ShowIcons {
		textX = itemX
	}
	
	for _, d := range data {
		if !d.Show {
			continue
		}
		
		// Icon
		if options.ShowIcons {
			sb.WriteString(fmt.Sprintf(`<g transform="translate(%d, %d)">`, iconX, itemY-10))
			sb.WriteString(getIconSVG(d.Icon, options.Theme.IconColor))
			sb.WriteString(`</g>`)
		}
		
		// Label
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s">%s:</text>`,
			textX, itemY, options.Theme.TextColor, d.Name))
		
		// Value
		value := utils.FormatNumber(d.Value, options.NumberFormat, options.NumberPrecision)
		fontWeight := "normal"
		if options.TextBold {
			fontWeight = "bold"
		}
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" font-weight="%s" fill="#%s" text-anchor="end">%s</text>`,
			cardWidth-130, itemY, fontWeight, options.Theme.TextColor, value))
		
		itemY += options.LineHeight
	}
	
	// Rank circle
	if !options.HideRank {
		centerX := cardWidth - 55
		centerY := 55 + (visibleCount*options.LineHeight)/2
		radius := 35
		strokeWidth := 5
		
		// Background circle
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="none" stroke="#%s" stroke-width="%d" opacity="0.2"/>`,
			centerX, centerY, radius, options.Theme.RingColor, strokeWidth))
		
		// Progress circle
		circumference := 2 * math.Pi * float64(radius)
		progress := (100 - percentile) / 100
		dashArray := circumference * progress
		dashOffset := circumference - dashArray
		
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="none" stroke="#%s" stroke-width="%d" 
			stroke-dasharray="%.2f %.2f" stroke-dashoffset="%.2f" stroke-linecap="round"
			transform="rotate(-90 %d %d)"/>`,
			centerX, centerY, radius, options.Theme.RingColor, strokeWidth,
			dashArray, circumference-dashArray, dashOffset,
			centerX, centerY))
		
		// Rank text
		fontSize := "24"
		if len(rank) > 1 {
			fontSize = "20"
		}
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="%s" font-weight="bold" fill="#%s" text-anchor="middle" dominant-baseline="middle">%s</text>`,
			centerX, centerY, fontSize, options.Theme.TextColor, rank))
		
		// Percentile text
		if options.RankIcon == "percentile" {
			sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="10" fill="#%s" text-anchor="middle">Top %.0f%%</text>`,
				centerX, centerY+radius+15, options.Theme.TextColor, percentile))
		}
	}
	
	sb.WriteString(`</svg>`)
	return sb.String()
}

// langData holds language statistics for rendering
type langData struct {
	name       string
	bytes      int
	repos      int
	color      string
	score      float64
	percentage float64
}

// RenderLangsCard renders the top languages card
type LangsCardOptions struct {
	Username          string
	Hide              []string
	Theme             themes.Theme
	CustomTitle       string
	CardWidth         int
	HideBorder        bool
	BorderRadius      float64
	Layout            string // normal, compact, donut, donut-vertical, pie
	LangsCount        int
	HideTitle         bool
	HideProgress      bool
	StatsFormat       string // percentages, bytes
	SizeWeight        float64
	CountWeight       float64
	DisableAnimations bool
}

// RenderTopLangsCard renders the top languages SVG card
func RenderTopLangsCard(languages map[string]*github.LanguageStats, options LangsCardOptions) string {
	// Calculate weighted scores
	var langs []langData
	var totalScore float64
	
	for name, stats := range languages {
		if utils.Contains(options.Hide, name) {
			continue
		}
		
		// Apply weight algorithm
		score := math.Pow(float64(stats.Bytes), options.SizeWeight) * 
				 math.Pow(float64(stats.Repos), options.CountWeight)
		
		langs = append(langs, langData{
			name:  name,
			bytes: stats.Bytes,
			repos: stats.Repos,
			color: stats.Color,
			score: score,
		})
		totalScore += score
	}
	
	// Calculate percentages
	if totalScore > 0 {
		for i := range langs {
			langs[i].percentage = (langs[i].score / totalScore) * 100
		}
	}
	
	// Sort by percentage descending
	sort.Slice(langs, func(i, j int) bool {
		return langs[i].percentage > langs[j].percentage
	})
	
	// Limit languages
	if len(langs) > options.LangsCount {
		langs = langs[:options.LangsCount]
	}
	
	// Recalculate percentages for displayed languages
	var displayedTotal float64
	for _, l := range langs {
		displayedTotal += l.score
	}
	if displayedTotal > 0 {
		for i := range langs {
			langs[i].percentage = (langs[i].score / displayedTotal) * 100
		}
	}
	
	// Calculate dimensions
	cardWidth := options.CardWidth
	if cardWidth == 0 {
		cardWidth = 300
	}
	
	height := 200
	title := options.CustomTitle
	if title == "" {
		title = "Most Used Languages"
	}
	
	// Build SVG
	var sb strings.Builder
	
	bgColor := "#" + options.Theme.BGColor
	
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, 
		cardWidth, height, cardWidth, height))
	
	// Background
	borderProps := ""
	if !options.HideBorder {
		borderProps = fmt.Sprintf(` stroke="#%s" stroke-width="1"`, options.Theme.BorderColor)
	}
	sb.WriteString(fmt.Sprintf(`<rect x="0.5" y="0.5" width="%d" height="%d" rx="%.1f" fill="%s"%s/>`,
		cardWidth-1, height-1, options.BorderRadius, bgColor, borderProps))
	
	// Title
	if !options.HideTitle {
		sb.WriteString(fmt.Sprintf(`<text x="25" y="35" font-family="Segoe UI, Ubuntu, sans-serif" font-size="18" font-weight="bold" fill="#%s">%s</text>`,
			options.Theme.TitleColor, utils.EscapeXML(title)))
	}
	
	// Render based on layout
	switch options.Layout {
	case "compact":
		sb.WriteString(renderCompactLayout(langs, cardWidth, height, options))
	case "donut":
		sb.WriteString(renderDonutLayout(langs, cardWidth, height, options))
	case "donut-vertical":
		sb.WriteString(renderDonutVerticalLayout(langs, cardWidth, height, options))
	case "pie":
		sb.WriteString(renderPieLayout(langs, cardWidth, height, options))
	default:
		sb.WriteString(renderNormalLayout(langs, cardWidth, height, options))
	}
	
	sb.WriteString(`</svg>`)
	return sb.String()
}

func renderNormalLayout(langs []langData, cardWidth, height int, options LangsCardOptions) string {
	var sb strings.Builder
	
	// Handle empty data
	if len(langs) == 0 {
		centerX := cardWidth / 2
		centerY := height/2 + 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s" text-anchor="middle">No languages found</text>`,
			centerX, centerY, options.Theme.TextColor))
		return sb.String()
	}
	
	startY := 60
	itemHeight := 32
	barHeight := 8
	
	for i, l := range langs {
		if i >= options.LangsCount {
			break
		}
		
		color := l.color
		if color == "" {
			color = "#858585"
		}
		
		y := startY + i*itemHeight
		
		// Language color dot (center aligned with text)
		sb.WriteString(fmt.Sprintf(`<circle cx="25" cy="%d" r="4" fill="%s"/>`, y-3, color))
		
		// Language name (move up to avoid overlap with progress bar)
		sb.WriteString(fmt.Sprintf(`<text x="35" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s">%s</text>`,
			y-4, options.Theme.TextColor, l.name))
		
		// Progress bar
		if !options.HideProgress {
			barWidth := float64(cardWidth-80) * (l.percentage / 100)
			if math.IsNaN(barWidth) || barWidth < 0 {
				barWidth = 0
			}
			sb.WriteString(fmt.Sprintf(`<rect x="35" y="%d" width="%d" height="%d" rx="2" fill="#%s" opacity="0.2"/>`,
				y+4, cardWidth-80, barHeight, options.Theme.BorderColor))
			sb.WriteString(fmt.Sprintf(`<rect x="35" y="%d" width="%.0f" height="%d" rx="2" fill="%s"/>`,
				y+4, barWidth, barHeight, color))
		}
		
		// Percentage (align with language name)
		var valueStr string
		if options.StatsFormat == "bytes" {
			valueStr = formatBytes(l.bytes)
		} else {
			valueStr = fmt.Sprintf("%.1f%%", l.percentage)
		}
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s" text-anchor="end">%s</text>`,
			cardWidth-25, y-4, options.Theme.TextColor, valueStr))
	}
	
	return sb.String()
}

func renderCompactLayout(langs []langData, cardWidth, height int, options LangsCardOptions) string {
	var sb strings.Builder
	
	// Handle empty data
	if len(langs) == 0 {
		centerX := cardWidth / 2
		centerY := height/2 + 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s" text-anchor="middle">No languages found</text>`,
			centerX, centerY, options.Theme.TextColor))
		return sb.String()
	}
	
	startY := 55
	itemHeight := 22
	
	// Calculate items per row (2 items per row for compact layout)
	itemsPerRow := 2
	colWidth := (cardWidth - 50) / itemsPerRow
	
	for i, l := range langs {
		if i >= options.LangsCount {
			break
		}
		
		color := l.color
		if color == "" {
			color = "#858585"
		}
		
		row := i / itemsPerRow
		col := i % itemsPerRow
		x := 25 + col*colWidth
		y := startY + row*itemHeight
		
		// Language color dot
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="3" fill="%s"/>`, x, y, color))
		
		// Language name
		nameX := x + 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
			nameX, y+1, options.Theme.TextColor, l.name))
		
		// Percentage (right aligned in column)
		var valueStr string
		if options.StatsFormat == "bytes" {
			valueStr = formatBytes(l.bytes)
		} else {
			valueStr = fmt.Sprintf("%.1f%%", l.percentage)
		}
		percentX := x + colWidth - 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="11" fill="#%s" text-anchor="end">%s</text>`,
			percentX, y+1, options.Theme.TextColor, valueStr))
	}
	
	return sb.String()
}

func renderDonutLayout(langs []langData, cardWidth, height int, options LangsCardOptions) string {
	var sb strings.Builder
	
	// Handle empty data
	if len(langs) == 0 {
		centerX := cardWidth / 2
		centerY := height/2 + 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s" text-anchor="middle">No languages found</text>`,
			centerX, centerY, options.Theme.TextColor))
		return sb.String()
	}
	
	centerX := cardWidth / 2
	centerY := height/2 + 5
	radius := 50
	strokeWidth := 12
	
	// Render donut chart using SVG circle with stroke-dasharray
	circumference := 2 * math.Pi * float64(radius)
	currentAngle := -90.0 // Start from top
	
	for _, l := range langs {
		if l.percentage <= 0 || math.IsNaN(l.percentage) {
			continue
		}
		
		color := l.color
		if color == "" {
			color = "#858585"
		}
		
		// Calculate the length of the arc for this language
		arcLength := circumference * (l.percentage / 100)
		gapLength := circumference - arcLength
		
		// Calculate rotation to position this segment
		rotation := currentAngle
		
		// Create the circle segment
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="none" stroke="%s" stroke-width="%d"
			stroke-dasharray="%.2f %.2f" stroke-linecap="butt"
			transform="rotate(%.2f %d %d)"/>`,
			centerX, centerY, radius, color, strokeWidth,
			arcLength, gapLength,
			rotation, centerX, centerY))
		
		// Update angle for next segment
		currentAngle += (l.percentage / 100) * 360
	}
	
	// Center text showing total languages count
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="10" fill="#%s" text-anchor="middle" opacity="0.6">Total</text>`,
		centerX, centerY-5, options.Theme.TextColor))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="16" font-weight="bold" fill="#%s" text-anchor="middle">%d</text>`,
		centerX, centerY+15, options.Theme.TextColor, len(langs)))
	
	return sb.String()
}

func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

// RenderRepoCard renders a repository pin card
type RepoCardOptions struct {
	Theme             themes.Theme
	CustomTitle       string
	CardWidth         int
	HideBorder        bool
	BorderRadius      float64
	ShowOwner         bool
	DescriptionLines  int
	DisableAnimations bool
}

// RenderRepoCard renders a repository SVG card
func RenderRepoCard(repo *github.Repository, options RepoCardOptions) string {
	cardWidth := options.CardWidth
	if cardWidth == 0 {
		cardWidth = 400
	}
	
	height := 120
	if repo.Description != "" {
		height = 140
	}
	
	// Build SVG
	var sb strings.Builder
	
	bgColor := "#" + options.Theme.BGColor
	
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, 
		cardWidth, height, cardWidth, height))
	
	// Background
	borderProps := ""
	if !options.HideBorder {
		borderProps = fmt.Sprintf(` stroke="#%s" stroke-width="1"`, options.Theme.BorderColor)
	}
	sb.WriteString(fmt.Sprintf(`<rect x="0.5" y="0.5" width="%d" height="%d" rx="%.1f" fill="%s"%s/>`,
		cardWidth-1, height-1, options.BorderRadius, bgColor, borderProps))
	
	// Repo icon
	sb.WriteString(fmt.Sprintf(`<g transform="translate(25, 30)">`))
	sb.WriteString(getRepoIcon(options.Theme.TextColor))
	sb.WriteString(`</g>`)
	
	// Repo name
	title := repo.Name
	if options.ShowOwner {
		title = repo.FullName
	}
	sb.WriteString(fmt.Sprintf(`<text x="50" y="32" font-family="Segoe UI, Ubuntu, sans-serif" font-size="16" font-weight="bold" fill="#%s">%s</text>`,
		options.Theme.TitleColor, utils.EscapeXML(title)))
	
	// Description
	if repo.Description != "" {
		desc := utils.Truncate(repo.Description, 100)
		sb.WriteString(fmt.Sprintf(`<text x="25" y="60" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
			options.Theme.TextColor, utils.EscapeXML(desc)))
	}
	
	// Language
	langY := height - 35
	if repo.Language != "" {
		color := github.GetLanguageColor(repo.Language)
		sb.WriteString(fmt.Sprintf(`<circle cx="30" cy="%d" r="4" fill="%s"/>`, langY, color))
		sb.WriteString(fmt.Sprintf(`<text x="40" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
			langY+1, options.Theme.TextColor, repo.Language))
	}
	
	// Stars
	starX := cardWidth - 100
	sb.WriteString(fmt.Sprintf(`<g transform="translate(%d, %d)">`, starX, langY-8))
	sb.WriteString(getStarIcon(options.Theme.TextColor))
	sb.WriteString(`</g>`)
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
		starX+20, langY+1, options.Theme.TextColor, utils.FormatNumber(repo.Stars, "short", 1)))
	
	// Forks
	forkX := cardWidth - 50
	sb.WriteString(fmt.Sprintf(`<g transform="translate(%d, %d)">`, forkX, langY-8))
	sb.WriteString(getForkIcon(options.Theme.TextColor))
	sb.WriteString(`</g>`)
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
		forkX+20, langY+1, options.Theme.TextColor, utils.FormatNumber(repo.Forks, "short", 1)))
	
	sb.WriteString(`</svg>`)
	return sb.String()
}

// RenderWakaTimeCard renders a WakaTime stats card
type WakaTimeCardOptions struct {
	Theme             themes.Theme
	CustomTitle       string
	CardWidth         int
	HideBorder        bool
	BorderRadius      float64
	HideTitle         bool
	HideProgress      bool
	Layout            string // default, compact
	LangsCount        int
	LineHeight        int
	DisableAnimations bool
}

// RenderWakaTimeCard renders the WakaTime SVG card
func RenderWakaTimeCard(stats map[string]int, options WakaTimeCardOptions) string {
	cardWidth := options.CardWidth
	if cardWidth == 0 {
		cardWidth = 495
	}
	
	title := options.CustomTitle
	if title == "" {
		title = "WakaTime Stats"
	}
	
	height := 150
	if !options.HideTitle {
		height += 40
	}
	
	// Build SVG
	var sb strings.Builder
	
	bgColor := "#" + options.Theme.BGColor
	
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, 
		cardWidth, height, cardWidth, height))
	
	// Background
	borderProps := ""
	if !options.HideBorder {
		borderProps = fmt.Sprintf(` stroke="#%s" stroke-width="1"`, options.Theme.BorderColor)
	}
	sb.WriteString(fmt.Sprintf(`<rect x="0.5" y="0.5" width="%d" height="%d" rx="%.1f" fill="%s"%s/>`,
		cardWidth-1, height-1, options.BorderRadius, bgColor, borderProps))
	
	// Title
	if !options.HideTitle {
		sb.WriteString(fmt.Sprintf(`<text x="25" y="35" font-family="Segoe UI, Ubuntu, sans-serif" font-size="18" font-weight="bold" fill="#%s">%s</text>`,
			options.Theme.TitleColor, utils.EscapeXML(title)))
	}
	
	// Stats content
	startY := 70
	if options.HideTitle {
		startY = 30
	}
	
	sb.WriteString(fmt.Sprintf(`<text x="25" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s">Coming soon...</text>`,
		startY, options.Theme.TextColor))
	
	sb.WriteString(`</svg>`)
	return sb.String()
}

// Helper functions for icons
func getIconSVG(icon, color string) string {
	// Simplified icons as SVG paths
	switch icon {
	case "★":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M8 0l2 5h5l-4 3 2 5-5-3-5 3 2-5-4-3h5z"/></svg>`, color)
	case "📝":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M3 1h10v14H3V1zm1 1v12h8V2H4z"/><path d="M5 4h6v1H5zm0 2h6v1H5zm0 2h4v1H5z"/></svg>`, color)
	case "🔀":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M5 3a2 2 0 00-2 2v2H1v2h2v2a2 2 0 002 2h2v-2H5V5h2V3H5zm6 0v2h2v6h-2v2h2a2 2 0 002-2V5a2 2 0 00-2-2h-2z"/></svg>`, color)
	case "●":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><circle cx="8" cy="8" r="6"/></svg>`, color)
	case "📂":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M1 3h5l2 2h7v8H1V3z"/></svg>`, color)
	case "👀":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M8 3c-3.5 0-6.5 3-8 5 1.5 2 4.5 5 8 5s6.5-3 8-5c-1.5-2-4.5-5-8-5zm0 8a3 3 0 110-6 3 3 0 010 6z"/></svg>`, color)
	case "💬":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M1 3h14v8H8l-3 3v-3H1V3z"/></svg>`, color)
	case "✓":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><path d="M13.5 3.5L6 11 2.5 7.5l-1 1L6 13l8.5-8.5-1-1z"/></svg>`, color)
	case "%":
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><text x="3" y="13" font-size="12">%%</text></svg>`, color)
	default:
		return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s"><circle cx="8" cy="8" r="4"/></svg>`, color)
	}
}

func getRepoIcon(color string) string {
	return fmt.Sprintf(`<svg width="16" height="16" viewBox="0 0 16 16" fill="#%s">
		<path fill-rule="evenodd" d="M2 2.5A2.5 2.5 0 014.5 0h8.75a.75.75 0 01.75.75v12.5a.75.75 0 01-.75.75h-2.5a.75.75 0 110-1.5h1.75v-2h-8a1 1 0 00-.714 1.7.75.75 0 01-1.072 1.05A2.495 2.495 0 012 11.5v-9zm10.5-1V9h-8c-.356 0-.694.074-1 .208V2.5a1 1 0 011-1h8zM5 12.25v3.25a.25.25 0 00.4.2l1.45-1.087a.25.25 0 01.3 0L8.6 15.7a.25.25 0 00.4-.2v-3.25a.25.25 0 00-.25-.25h-3.5a.25.25 0 00-.25.25z"/>
	</svg>`, color)
}

func getStarIcon(color string) string {
	return fmt.Sprintf(`<svg width="14" height="14" viewBox="0 0 16 16" fill="#%s">
		<path d="M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.612a.75.75 0 01.416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3.766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.374a.75.75 0 01.416-1.28l4.21-.611L7.327.668A.75.75 0 018 .25z"/>
	</svg>`, color)
}

func getForkIcon(color string) string {
	return fmt.Sprintf(`<svg width="14" height="14" viewBox="0 0 16 16" fill="#%s">
		<path d="M5 3.25a1.25 1.25 0 11-2.5 0 1.25 1.25 0 012.5 0zm0 2.75a1.25 1.25 0 11-2.5 0 1.25 1.25 0 012.5 0zm0 2.75a1.25 1.25 0 11-2.5 0 1.25 1.25 0 012.5 0z"/>
		<path fill-rule="evenodd" d="M8 1.5A1.75 1.75 0 006.25 3v4.5a.25.25 0 01-.25.25H3.75a.25.25 0 00-.25.25v3.5c0 .138.112.25.25.25h2.25a.25.25 0 01.25.25V15a.75.75 0 001.5 0v-2.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V15a.75.75 0 001.5 0v-2.75a.25.25 0 01.25-.25h2.25a.25.25 0 00.25-.25v-3.5a.25.25 0 00-.25-.25h-2.25a.25.25 0 01-.25-.25V3A1.75 1.75 0 008 1.5z"/>
	</svg>`, color)
}
