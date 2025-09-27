package themes

type Theme struct {
	Name              string
	GlamourStyle      string
	TitleColor        string
	TitleColorFg      string
	SelectedItemColor string
	FilterColor       string
	HighlightStyle    string // "background", "underline", "prefix", "prefix-underline"
}

var AvailableThemes = []Theme{
	{
		Name:              "dark",
		GlamourStyle:      "dark",
		TitleColor:        "62",
		TitleColorFg:      "231",
		SelectedItemColor: "170",
		FilterColor:       "#555555",
		HighlightStyle:    "prefix-underline",
	},
	{
		Name:              "light",
		GlamourStyle:      "light",
		TitleColor:        "12",
		TitleColorFg:      "0",
		SelectedItemColor: "75",
		FilterColor:       "#999999",
		HighlightStyle:    "prefix-underline",
	},
	{
		Name:              "dracula",
		GlamourStyle:      "dracula",
		TitleColor:        "141",
		TitleColorFg:      "231",
		SelectedItemColor: "212",
		FilterColor:       "#6272a4",
		HighlightStyle:    "prefix-underline",
	},
	{
		Name:              "pink",
		GlamourStyle:      "pink",
		TitleColor:        "200",
		TitleColorFg:      "0",
		SelectedItemColor: "205",
		FilterColor:       "#cc99cc",
		HighlightStyle:    "prefix-underline",
	},
	{
		Name:              "ascii",
		GlamourStyle:      "ascii",
		TitleColor:        "7",
		TitleColorFg:      "0",
		SelectedItemColor: "7",
		FilterColor:       "#808080",
		HighlightStyle:    "prefix",
	},
}

func GetThemeByName(name string) *Theme {
	for i := range AvailableThemes {
		if AvailableThemes[i].Name == name {
			return &AvailableThemes[i]
		}
	}
	// Return default dark theme if not found
	return &AvailableThemes[0]
}

func GetThemeNames() []string {
	names := make([]string, len(AvailableThemes))
	for i, theme := range AvailableThemes {
		names[i] = theme.Name
	}
	return names
}

func GetHighlightStyles() []string {
	return []string{
		"background",
		"underline",
		"prefix",
		"prefix-underline",
	}
}

func GetSpinnerTypes() []string {
	return []string{
		"braille",
		"dots",
		"line",
		"arrow",
		"star",
		"circle",
		"square",
		"triangle",
	}
}

func GetSpinnerFrames(spinnerType string) []string {
	switch spinnerType {
	case "braille":
		return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	case "dots":
		return []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	case "line":
		return []string{"-", "\\", "|", "/"}
	case "arrow":
		return []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	case "star":
		return []string{"✶", "✸", "✹", "✺", "✹", "✸"}
	case "circle":
		return []string{"◐", "◓", "◑", "◒"}
	case "square":
		return []string{"◰", "◳", "◲", "◱"}
	case "triangle":
		return []string{"◢", "◣", "◤", "◥"}
	default:
		// Default to braille
		return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
}