package cards

import (
	"fmt"
	"math"
	"strings"
)

// renderDonutVerticalLayout renders a vertical donut chart with legend on the side
func renderDonutVerticalLayout(langs []langData, cardWidth, height int, options LangsCardOptions) string {
	var sb strings.Builder
	if len(langs) == 0 {
		centerX := cardWidth / 2
		centerY := height/2 + 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s" text-anchor="middle">No languages found</text>`,
			centerX, centerY, options.Theme.TextColor))
		return sb.String()
	}

	donutCenterX := 75
	donutCenterY := height/2 + 5
	radius := 42
	strokeWidth := 10
	circumference := 2 * math.Pi * float64(radius)
	currentAngle := -90.0

	for _, l := range langs {
		if l.percentage <= 0 || math.IsNaN(l.percentage) {
			continue
		}
		color := l.color
		if color == "" {
			color = "#858585"
		}
		arcLength := circumference * (l.percentage / 100)
		gapLength := circumference - arcLength
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="none" stroke="%s" stroke-width="%d" stroke-dasharray="%.2f %.2f" stroke-linecap="butt" transform="rotate(%.2f %d %d)"/>`,
			donutCenterX, donutCenterY, radius, color, strokeWidth, arcLength, gapLength, currentAngle, donutCenterX, donutCenterY))
		currentAngle += (l.percentage / 100) * 360
	}

	sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="10" fill="#%s" text-anchor="middle" opacity="0.6">Total</text>`,
		donutCenterX, donutCenterY-3, options.Theme.TextColor))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" font-weight="bold" fill="#%s" text-anchor="middle">%d</text>`,
		donutCenterX, donutCenterY+12, options.Theme.TextColor, len(langs)))

	legendX := 140
	startY := 60
	itemHeight := 22
	for i, l := range langs {
		if i >= options.LangsCount {
			break
		}
		color := l.color
		if color == "" {
			color = "#858585"
		}
		y := startY + i*itemHeight
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="%s"/>`, legendX, y, color))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
			legendX+12, y+1, options.Theme.TextColor, l.name))
		var valueStr string
		if options.StatsFormat == "bytes" {
			valueStr = formatBytes(l.bytes)
		} else {
			valueStr = fmt.Sprintf("%.1f%%", l.percentage)
		}
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="11" fill="#%s" text-anchor="end">%s</text>`,
			cardWidth-20, y+1, options.Theme.TextColor, valueStr))
	}
	return sb.String()
}

// renderPieLayout renders a pie chart with legend
func renderPieLayout(langs []langData, cardWidth, height int, options LangsCardOptions) string {
	var sb strings.Builder
	if len(langs) == 0 {
		centerX := cardWidth / 2
		centerY := height/2 + 10
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="14" fill="#%s" text-anchor="middle">No languages found</text>`,
			centerX, centerY, options.Theme.TextColor))
		return sb.String()
	}

	pieCenterX := 75
	pieCenterY := height/2 + 5
	radius := 48.0
	currentAngle := -90.0

	for _, l := range langs {
		if l.percentage <= 0 || math.IsNaN(l.percentage) {
			continue
		}
		color := l.color
		if color == "" {
			color = "#858585"
		}
		startAngle := currentAngle
		endAngle := currentAngle + (l.percentage/100)*360
		startRad := startAngle * math.Pi / 180
		endRad := endAngle * math.Pi / 180
		x1 := float64(pieCenterX) + radius*math.Cos(startRad)
		y1 := float64(pieCenterY) + radius*math.Sin(startRad)
		x2 := float64(pieCenterX) + radius*math.Cos(endRad)
		y2 := float64(pieCenterY) + radius*math.Sin(endRad)
		largeArc := 0
		if l.percentage > 50 {
			largeArc = 1
		}
		sb.WriteString(fmt.Sprintf(`<path d="M %d %d L %.2f %.2f A %.0f %.0f 0 %d 1 %.2f %.2f Z" fill="%s"/>`,
			pieCenterX, pieCenterY, x1, y1, radius, radius, largeArc, x2, y2, color))
		currentAngle = endAngle
	}

	legendX := 140
	startY := 60
	itemHeight := 22
	for i, l := range langs {
		if i >= options.LangsCount {
			break
		}
		color := l.color
		if color == "" {
			color = "#858585"
		}
		y := startY + i*itemHeight
		sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="%s"/>`, legendX, y, color))
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="12" fill="#%s">%s</text>`,
			legendX+12, y+1, options.Theme.TextColor, l.name))
		var valueStr string
		if options.StatsFormat == "bytes" {
			valueStr = formatBytes(l.bytes)
		} else {
			valueStr = fmt.Sprintf("%.1f%%", l.percentage)
		}
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Segoe UI, Ubuntu, sans-serif" font-size="11" fill="#%s" text-anchor="end">%s</text>`,
			cardWidth-20, y+1, options.Theme.TextColor, valueStr))
	}
	return sb.String()
}
