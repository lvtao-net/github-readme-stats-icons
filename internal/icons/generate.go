package icons

import (
	"encoding/json"
	"os"
)

// IconInfo holds information about an icon
type IconInfo struct {
	Name      string `json:"name"`
	HasDark   bool   `json:"hasDark"`
	HasLight  bool   `json:"hasLight"`
	ShortName string `json:"shortName,omitempty"`
}

// IconsData holds all icons metadata
type IconsData struct {
	Icons       []IconInfo        `json:"icons"`
	ShortNames  map[string]string `json:"shortNames"`
	ThemedIcons []string          `json:"themedIcons"`
	Total       int               `json:"total"`
}

// GenerateIconsJSON generates icons.json file
func (m *Manager) GenerateIconsJSON(outputPath string) error {
	allIcons := m.GetAllIcons()
	
	data := IconsData{
		Icons:       make([]IconInfo, 0, len(allIcons)),
		ShortNames:  m.shortNames,
		ThemedIcons: make([]string, 0),
		Total:       len(allIcons),
	}
	
	// Build reverse mapping (full name -> short name)
	reverseShortNames := make(map[string]string)
	for short, full := range m.shortNames {
		reverseShortNames[full] = short
	}
	
	// Build themed icons list
	for iconName := range m.themedIcons {
		data.ThemedIcons = append(data.ThemedIcons, iconName)
	}
	
	// Build icon info list
	for _, name := range allIcons {
		info := IconInfo{
			Name: name,
		}
		
		// Check for theme variants
		if m.themedIcons[name] {
			info.HasDark = true
			info.HasLight = true
		}
		
		// Check for short name
		if short, ok := reverseShortNames[name]; ok {
			info.ShortName = short
		}
		
		data.Icons = append(data.Icons, info)
	}
	
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(outputPath, jsonData, 0644)
}

// LoadIconsJSON loads icons data from JSON file
func LoadIconsJSON(path string) (*IconsData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var iconsData IconsData
	if err := json.Unmarshal(data, &iconsData); err != nil {
		return nil, err
	}
	
	return &iconsData, nil
}

// GetIconsList returns a simple list of icon names
func (m *Manager) GetIconsList() []string {
	return m.GetAllIcons()
}

// GetIconsMetadata returns detailed metadata for all icons
func (m *Manager) GetIconsMetadata() IconsData {
	allIcons := m.GetAllIcons()
	
	data := IconsData{
		Icons:       make([]IconInfo, 0, len(allIcons)),
		ShortNames:  m.shortNames,
		ThemedIcons: make([]string, 0),
		Total:       len(allIcons),
	}
	
	// Build reverse mapping
	reverseShortNames := make(map[string]string)
	for short, full := range m.shortNames {
		reverseShortNames[full] = short
	}
	
	// Build themed icons list
	for iconName := range m.themedIcons {
		data.ThemedIcons = append(data.ThemedIcons, iconName)
	}
	
	// Build icon info
	for _, name := range allIcons {
		info := IconInfo{
			Name: name,
		}
		
		if m.themedIcons[name] {
			info.HasDark = true
			info.HasLight = true
		}
		
		if short, ok := reverseShortNames[name]; ok {
			info.ShortName = short
		}
		
		data.Icons = append(data.Icons, info)
	}
	
	return data
}
