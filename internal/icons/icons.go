package icons

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github-readme-stats/internal/config"
)

// Manager handles skill icons
type Manager struct {
	icons       map[string]string // name -> SVG content
	themedIcons map[string]bool   // icons with light/dark variants
	shortNames  map[string]string // short name -> full name
	mu          sync.RWMutex
}

// shortNamesMapping maps short names to full icon names (case-insensitive matching to actual filenames)
var shortNamesMapping = map[string]string{
	"js":         "javascript",
	"ts":         "typescript",
	"py":         "python",
	"tailwind":   "tailwindcss",
	"vue":        "vuejs",
	"nuxt":       "nuxtjs",
	"go":         "golang",
	"cf":         "cloudflare",
	"wasm":       "webassembly",
	"postgres":   "postgresql",
	"k8s":        "kubernetes",
	"next":       "nextjs",
	"mongo":      "mongodb",
	"md":         "markdown",
	"ps":         "photoshop",
	"ai":         "illustrator",
	"pr":         "premiere",
	"ae":         "aftereffects",
	"scss":       "sass",
	"sc":         "scala",
	"net":        "dotnet",
	"gatsbyjs":   "gatsby",
	"gql":        "graphql",
	"vlang":      "v",
	"aws":        "aws",
	"bots":       "discordbots",
	"express":    "expressjs",
	"gcp":        "gcp",
	"mui":        "materialui",
	"windi":      "windicss",
	"unreal":     "unrealengine",
	"nest":       "nestjs",
	"ktorio":     "ktor",
	"pwsh":       "powershell",
	"au":         "audition",
	"rollup":     "rollupjs",
	"rxjs":       "reactivex",
	"rxjava":     "reactivex",
	"ghactions":  "githubactions",
	"sklearn":    "scikitlearn",
	"cs":         "cs",
	"cpp":        "cpp",
	"php":        "php",
	"html":       "html",
	"css":        "css",
	"java":       "java",
	"linux":      "linux",
	"nginx":      "nginx",
	"mysql":      "mysql",
	"redis":      "redis",
	"sqlite":     "sqlite",
	"vscode":     "vscode",
	"vim":        "vim",
	"git":        "git",
	"c":          "c",
	"laravel":    "laravel",
	"wordpress":  "wordpress",
	"docker":     "docker",
	"kubernetes": "kubernetes",
	"react":      "react",
	"angular":    "angular",
	"svelte":     "svelte",
	"node":       "nodejs",
	"rust":       "rust",
	"ruby":       "ruby",
	"swift":      "swift",
	"kotlin":     "kotlin",
}

// NewManager creates a new icon manager
func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		icons:       make(map[string]string),
		themedIcons: make(map[string]bool),
		shortNames:  shortNamesMapping,
	}

	// Load icons from assets directory
	m.loadIcons(cfg.AssetsPath)

	return m
}

// loadIcons loads all SVG icons from the assets directory
func (m *Manager) loadIcons(assetsPath string) {
	if assetsPath == "" {
		assetsPath = "assets/icons"
	}

	entries, err := os.ReadDir(assetsPath)
	if err != nil {
		return
	}

	// First pass: identify themed icons and load all icons
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".svg") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".svg")
		lowerName := strings.ToLower(name)

		// Check if this is a themed icon (has -light or -dark suffix)
		baseName := lowerName
		if strings.HasSuffix(lowerName, "-light") {
			baseName = strings.TrimSuffix(lowerName, "-light")
			m.themedIcons[baseName] = true
		} else if strings.HasSuffix(lowerName, "-dark") {
			baseName = strings.TrimSuffix(lowerName, "-dark")
			m.themedIcons[baseName] = true
		}

		// Read SVG content
		data, err := os.ReadFile(filepath.Join(assetsPath, entry.Name()))
		if err != nil {
			continue
		}

		// Store with lowercase key
		m.mu.Lock()
		m.icons[lowerName] = string(data)
		m.mu.Unlock()
	}
}

// GetIcon returns an icon SVG by name
func (m *Manager) GetIcon(name string, theme string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Resolve short name to full name
	fullName := name
	if resolved, ok := m.shortNames[strings.ToLower(name)]; ok {
		fullName = resolved
	}
	fullName = strings.ToLower(fullName)

	// Check if this is a themed icon
	if m.themedIcons[fullName] {
		// Try theme-specific version first
		themedName := fullName + "-" + theme
		if svg, ok := m.icons[themedName]; ok {
			return svg, true
		}
		// Fall back to the other theme
		otherTheme := "light"
		if theme == "light" {
			otherTheme = "dark"
		}
		if svg, ok := m.icons[fullName+"-"+otherTheme]; ok {
			return svg, true
		}
	}

	// Try the plain name
	if svg, ok := m.icons[fullName]; ok {
		return svg, true
	}

	return "", false
}

// IsValidIcon checks if an icon exists
func (m *Manager) IsValidIcon(name string) bool {
	// Resolve short name
	fullName := name
	if resolved, ok := m.shortNames[strings.ToLower(name)]; ok {
		fullName = resolved
	}
	fullName = strings.ToLower(fullName)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check themed variants
	if _, ok := m.icons[fullName+"-dark"]; ok {
		return true
	}
	if _, ok := m.icons[fullName+"-light"]; ok {
		return true
	}
	_, ok := m.icons[fullName]
	return ok
}

// GetAllIcons returns all available icon names
func (m *Manager) GetAllIcons() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return unique base names (without theme suffixes)
	nameMap := make(map[string]bool)
	for name := range m.icons {
		// Skip theme-specific files, their base name will be added via the themedIcons map
		if strings.HasSuffix(name, "-light") || strings.HasSuffix(name, "-dark") {
			continue
		}
		nameMap[name] = true
	}

	// Also add themed icons' base names
	for baseName := range m.themedIcons {
		nameMap[baseName] = true
	}

	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}

	// Sort names for consistent ordering
	sort.Strings(names)

	return names
}

// GenerateSVG generates an SVG with multiple icons
func (m *Manager) GenerateSVG(iconNames []string, theme string, perLine int) string {
	if perLine < 1 {
		perLine = 15
	}
	if perLine > 50 {
		perLine = 50
	}

	// Icon size (48x48 like skillicons)
	const iconSize = 48
	const gap = 10
	// Source icon coordinate system (256x256 like skillicons)
	const sourceSize = 256

	// Calculate output dimensions
	numIcons := len(iconNames)
	numLines := (numIcons + perLine - 1) / perLine
	// ViewBox size (source coordinate system) - exact size without extra padding
	viewBoxWidth := perLine*sourceSize + (perLine-1)*gap
	viewBoxHeight := numLines*sourceSize + (numLines-1)*gap

	var sb strings.Builder
	// Outer SVG: use viewBox as width/height to ensure 1:1 scaling and left alignment
	sb.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`,
		viewBoxWidth, viewBoxHeight, viewBoxWidth, viewBoxHeight))

	// Add icons
	for i, name := range iconNames {
		if svgContent, ok := m.GetIcon(name, theme); ok {
			row := i / perLine
			col := i % perLine
			// Position in source coordinate system
			x := col * (sourceSize + gap)
			y := row * (sourceSize + gap)

			// Extract the icon with proper sizing and unique IDs
			iconSVG := prepareIcon(svgContent, i)

			sb.WriteString(fmt.Sprintf(`<g transform="translate(%d, %d)">%s</g>`, x, y, iconSVG))
		}
	}

	sb.WriteString(`</svg>`)
	return sb.String()
}

// prepareIcon prepares an icon SVG with proper sizing and unique IDs
func prepareIcon(svg string, index int) string {
	// Check if content uses xlink
	usesXLink := strings.Contains(svg, "xlink:") || strings.Contains(svg, "xlink:href")

	// Extract content between svg tags
	svgStart := strings.Index(svg, "<svg")
	if svgStart < 0 {
		return svg
	}

	// Find the end of opening <svg ...> tag
	start := svgStart
	inQuote := false
	quoteChar := byte(0)
	for i := svgStart + 4; i < len(svg); i++ {
		c := svg[i]
		if !inQuote {
			if c == '"' || c == '\'' {
				inQuote = true
				quoteChar = c
			} else if c == '>' {
				start = i
				break
			}
		} else {
			if c == quoteChar {
				inQuote = false
			}
		}
	}

	end := strings.LastIndex(svg, "</svg>")
	if start <= svgStart || end <= start {
		return svg
	}

	content := strings.TrimSpace(svg[start+1 : end])

	// Create unique prefix for this icon's IDs
	idPrefix := fmt.Sprintf("icon%d-", index)

	// Replace all IDs and references to make them unique
	content = regexp.MustCompile(`id="([^"]+)"`).ReplaceAllString(content, `id="`+idPrefix+`$1"`)
	content = regexp.MustCompile(`url\(#([^)]+)\)`).ReplaceAllString(content, `url(#`+idPrefix+`$1)`)
	content = regexp.MustCompile(`href="#([^"]+)"`).ReplaceAllString(content, `href="#`+idPrefix+`$1"`)
	// Also replace xlink:href references
	content = regexp.MustCompile(`xlink:href="#([^"]+)"`).ReplaceAllString(content, `xlink:href="#`+idPrefix+`$1"`)

	// Build the icon SVG with source size (256x256)
	// Add xmlns:xlink if the content uses xlink
	if usesXLink {
		return fmt.Sprintf(`<svg width="256" height="256" viewBox="0 0 256 256" xmlns:xlink="http://www.w3.org/1999/xlink">%s</svg>`, content)
	}
	return fmt.Sprintf(`<svg width="256" height="256" viewBox="0 0 256 256">%s</svg>`, content)
}




