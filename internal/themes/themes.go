package themes

// Theme represents a color theme for cards
type Theme struct {
	TitleColor  string
	TextColor   string
	IconColor   string
	BorderColor string
	BGColor     string
	RingColor   string
}

// Themes is a collection of available themes
var Themes = map[string]Theme{
	"default": {
		TitleColor:  "2f80ed",
		TextColor:   "434d58",
		IconColor:   "4c71f2",
		BorderColor: "e4e2e2",
		BGColor:     "fffefe",
		RingColor:   "2f80ed",
	},
	"dark": {
		TitleColor:  "fff",
		TextColor:   "9f9f9f",
		IconColor:   "79ff97",
		BorderColor: "e4e2e2",
		BGColor:     "151515",
		RingColor:   "bb8aff",
	},
	"radical": {
		TitleColor:  "fe428e",
		TextColor:   "a9fef7",
		IconColor:   "f8d847",
		BorderColor: "e4e2e2",
		BGColor:     "141321",
		RingColor:   "fe428e",
	},
	"merko": {
		TitleColor:  "abd200",
		TextColor:   "68b587",
		IconColor:   "b7d364",
		BorderColor: "e4e2e2",
		BGColor:     "0a0f0b",
		RingColor:   "abd200",
	},
	"gruvbox": {
		TitleColor:  "fabd2f",
		TextColor:   "8ec07c",
		IconColor:   "fe8019",
		BorderColor: "e4e2e2",
		BGColor:     "282828",
		RingColor:   "fabd2f",
	},
	"tokyonight": {
		TitleColor:  "70a5fd",
		TextColor:   "38bdae",
		IconColor:   "bf91f3",
		BorderColor: "e4e2e2",
		BGColor:     "1a1b27",
		RingColor:   "70a5fd",
	},
	"onedark": {
		TitleColor:  "e06c75",
		TextColor:   "abb2bf",
		IconColor:   "98c379",
		BorderColor: "e4e2e2",
		BGColor:     "282c34",
		RingColor:   "e06c75",
	},
	"cobalt": {
		TitleColor:  "e683d9",
		TextColor:   "75eeb2",
		IconColor:   "0480ef",
		BorderColor: "e4e2e2",
		BGColor:     "193549",
		RingColor:   "e683d9",
	},
	"synthwave": {
		TitleColor:  "e2e9ec",
		TextColor:   "e5289e",
		IconColor:   "ef8539",
		BorderColor: "e4e2e2",
		BGColor:     "2b213a",
		RingColor:   "e2e9ec",
	},
	"highcontrast": {
		TitleColor:  "e7f216",
		TextColor:   "fff",
		IconColor:   "00ffff",
		BorderColor: "e4e2e2",
		BGColor:     "000",
		RingColor:   "e7f216",
	},
	"dracula": {
		TitleColor:  "ff79c6",
		TextColor:   "f8f8f2",
		IconColor:   "8be9fd",
		BorderColor: "e4e2e2",
		BGColor:     "282a36",
		RingColor:   "ff79c6",
	},
	"prussian": {
		TitleColor:  "bddfff",
		TextColor:   "6e93b5",
		IconColor:   "38a0ff",
		BorderColor: "e4e2e2",
		BGColor:     "172f45",
		RingColor:   "bddfff",
	},
	"monokai": {
		TitleColor:  "eb1f6a",
		TextColor:   "f1f1eb",
		IconColor:   "e28905",
		BorderColor: "e4e2e2",
		BGColor:     "272822",
		RingColor:   "eb1f6a",
	},
	"vue": {
		TitleColor:  "41b883",
		TextColor:   "273849",
		IconColor:   "41b883",
		BorderColor: "e4e2e2",
		BGColor:     "fffefe",
		RingColor:   "41b883",
	},
	"vue-dark": {
		TitleColor:  "41b883",
		TextColor:   "fffefe",
		IconColor:   "41b883",
		BorderColor: "e4e2e2",
		BGColor:     "273849",
		RingColor:   "41b883",
	},
	"github-dark": {
		TitleColor:  "39d0d8",
		TextColor:   "c9d1d9",
		IconColor:   "58a6ff",
		BorderColor: "e4e2e2",
		BGColor:     "0d1117",
		RingColor:   "39d0d8",
	},
	"github-dark-blue": {
		TitleColor:  "39d0d8",
		TextColor:   "c9d1d9",
		IconColor:   "58a6ff",
		BorderColor: "e4e2e2",
		BGColor:     "000000",
		RingColor:   "39d0d8",
	},
	"transparent": {
		TitleColor:  "006aff",
		TextColor:   "417e87",
		IconColor:   "0579c3",
		BorderColor: "e4e2e2",
		BGColor:     "00000000",
		RingColor:   "006aff",
	},
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["default"]
}

// ParseColor parses a color string and returns proper format
func ParseColor(color string) string {
	// Remove # if present
	if len(color) > 0 && color[0] == '#' {
		color = color[1:]
	}
	return color
}
