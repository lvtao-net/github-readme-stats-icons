package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FormatNumber formats a number with k/m suffix
func FormatNumber(num int, format string, precision int) string {
	if format == "long" {
		return strconv.Itoa(num)
	}
	
	// Short format
	if num < 1000 {
		return strconv.Itoa(num)
	}
	
	suffix := "k"
	value := float64(num) / 1000
	
	if num >= 1000000 {
		suffix = "m"
		value = float64(num) / 1000000
	}
	
	// Handle invalid precision (default to 1 decimal place)
	if precision < 0 {
		precision = 1
	}
	
	if precision == 0 {
		return fmt.Sprintf("%d%s", int(value), suffix)
	}
	
	formatStr := fmt.Sprintf("%%.%df%%s", precision)
	return fmt.Sprintf(formatStr, value, suffix)
}

// Clamp restricts a value to a range
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ParseBool parses a boolean string
func ParseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes"
}

// ParseInt parses an integer with default
func ParseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// ParseFloat parses a float with default
func ParseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

// Split splits a string by comma
func Split(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// Contains checks if a string slice contains a value
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// CalculateRank calculates user rank based on stats
func CalculateRank(commits, prs, issues, reviews, stars, followers int) (string, float64) {
	// This is a simplified version of the rank calculation
	// The original uses exponential and log-normal distributions
	
	// Calculate percentile based on various stats
	allStats := []struct {
		value    int
		weight   float64
		maxValue float64
	}{
		{commits, 0.3, 10000},
		{prs, 0.2, 500},
		{issues, 0.1, 500},
		{reviews, 0.15, 500},
		{stars, 0.15, 1000},
		{followers, 0.1, 500},
	}
	
	var totalScore float64
	for _, stat := range allStats {
		if stat.maxValue > 0 {
			score := math.Min(float64(stat.value)/stat.maxValue, 1.0)
			totalScore += score * stat.weight
		}
	}
	
	// Calculate percentile (higher score = lower percentile)
	percentile := (1 - totalScore) * 100
	
	// Determine rank
	var rank string
	switch {
	case percentile <= 1:
		rank = "S+"
		percentile = 1
	case percentile <= 12.5:
		rank = "A+"
	case percentile <= 25:
		rank = "A"
	case percentile <= 37.5:
		rank = "A-"
	case percentile <= 50:
		rank = "B+"
	case percentile <= 62.5:
		rank = "B"
	case percentile <= 75:
		rank = "B-"
	case percentile <= 87.5:
		rank = "C+"
	default:
		rank = "C"
	}
	
	return rank, percentile
}

// EscapeXML escapes special XML characters
func EscapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// Truncate truncates a string with ellipsis
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// Percentage calculates percentage
func Percentage(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}
