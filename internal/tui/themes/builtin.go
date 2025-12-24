package themes

// CatppuccinMocha returns the Catppuccin Mocha theme
func CatppuccinMocha() *Theme {
	return &Theme{
		Name:        "catppuccin-mocha",
		Author:      "Catppuccin",
		Description: "Soothing pastel theme for the high-spirited",
		Colors: ThemeColors{
			Background:       "#1E1E2E",
			Foreground:       "#CDD6F4",
			Primary:          "#89B4FA",
			Secondary:        "#A6ADC8",
			Accent:           "#F5C2E7",
			Success:          "#A6E3A1",
			Warning:          "#F9E2AF",
			Error:            "#F38BA8",
			Info:             "#89DCEB",
			Border:           "#45475A",
			BorderFocused:    "#89B4FA",
			Selection:        "#313244",
			Muted:            "#6C7086",
			UserMessage:      "#89B4FA",
			AssistantMessage: "#A6E3A1",
			SystemMessage:    "#F9E2AF",
			Keyword:          "#CBA6F7",
			String:           "#A6E3A1",
			Number:           "#FAB387",
			Comment:          "#6C7086",
			Function:         "#89B4FA",
		},
	}
}

// CatppuccinLatte returns the Catppuccin Latte (light) theme
func CatppuccinLatte() *Theme {
	return &Theme{
		Name:        "catppuccin-latte",
		Author:      "Catppuccin",
		Description: "Light variant of Catppuccin",
		Colors: ThemeColors{
			Background:       "#EFF1F5",
			Foreground:       "#4C4F69",
			Primary:          "#1E66F5",
			Secondary:        "#6C6F85",
			Accent:           "#EA76CB",
			Success:          "#40A02B",
			Warning:          "#DF8E1D",
			Error:            "#D20F39",
			Info:             "#04A5E5",
			Border:           "#BCC0CC",
			BorderFocused:    "#1E66F5",
			Selection:        "#CCD0DA",
			Muted:            "#9CA0B0",
			UserMessage:      "#1E66F5",
			AssistantMessage: "#40A02B",
			SystemMessage:    "#DF8E1D",
			Keyword:          "#8839EF",
			String:           "#40A02B",
			Number:           "#FE640B",
			Comment:          "#9CA0B0",
			Function:         "#1E66F5",
		},
	}
}

// Dracula returns the Dracula theme
func Dracula() *Theme {
	return &Theme{
		Name:        "dracula",
		Author:      "Dracula Theme",
		Description: "A dark theme for code editors",
		Colors: ThemeColors{
			Background:       "#282A36",
			Foreground:       "#F8F8F2",
			Primary:          "#BD93F9",
			Secondary:        "#6272A4",
			Accent:           "#FF79C6",
			Success:          "#50FA7B",
			Warning:          "#F1FA8C",
			Error:            "#FF5555",
			Info:             "#8BE9FD",
			Border:           "#44475A",
			BorderFocused:    "#BD93F9",
			Selection:        "#44475A",
			Muted:            "#6272A4",
			UserMessage:      "#8BE9FD",
			AssistantMessage: "#50FA7B",
			SystemMessage:    "#F1FA8C",
			Keyword:          "#FF79C6",
			String:           "#F1FA8C",
			Number:           "#BD93F9",
			Comment:          "#6272A4",
			Function:         "#50FA7B",
		},
	}
}

// Nord returns the Nord theme
func Nord() *Theme {
	return &Theme{
		Name:        "nord",
		Author:      "Arctic Ice Studio",
		Description: "An arctic, north-bluish color palette",
		Colors: ThemeColors{
			Background:       "#2E3440",
			Foreground:       "#ECEFF4",
			Primary:          "#88C0D0",
			Secondary:        "#81A1C1",
			Accent:           "#B48EAD",
			Success:          "#A3BE8C",
			Warning:          "#EBCB8B",
			Error:            "#BF616A",
			Info:             "#81A1C1",
			Border:           "#3B4252",
			BorderFocused:    "#88C0D0",
			Selection:        "#434C5E",
			Muted:            "#4C566A",
			UserMessage:      "#88C0D0",
			AssistantMessage: "#A3BE8C",
			SystemMessage:    "#EBCB8B",
			Keyword:          "#81A1C1",
			String:           "#A3BE8C",
			Number:           "#B48EAD",
			Comment:          "#616E88",
			Function:         "#88C0D0",
		},
	}
}

// TokyoNight returns the Tokyo Night theme
func TokyoNight() *Theme {
	return &Theme{
		Name:        "tokyo-night",
		Author:      "enkia",
		Description: "A clean, dark theme that celebrates Tokyo night life",
		Colors: ThemeColors{
			Background:       "#1A1B26",
			Foreground:       "#C0CAF5",
			Primary:          "#7AA2F7",
			Secondary:        "#9AA5CE",
			Accent:           "#BB9AF7",
			Success:          "#9ECE6A",
			Warning:          "#E0AF68",
			Error:            "#F7768E",
			Info:             "#7DCFFF",
			Border:           "#27293B",
			BorderFocused:    "#7AA2F7",
			Selection:        "#283457",
			Muted:            "#565F89",
			UserMessage:      "#7AA2F7",
			AssistantMessage: "#9ECE6A",
			SystemMessage:    "#E0AF68",
			Keyword:          "#BB9AF7",
			String:           "#9ECE6A",
			Number:           "#FF9E64",
			Comment:          "#565F89",
			Function:         "#7AA2F7",
		},
	}
}

// GruvboxDark returns the Gruvbox Dark theme
func GruvboxDark() *Theme {
	return &Theme{
		Name:        "gruvbox-dark",
		Author:      "morhetz",
		Description: "Retro groove color scheme",
		Colors: ThemeColors{
			Background:       "#282828",
			Foreground:       "#EBDBB2",
			Primary:          "#83A598",
			Secondary:        "#A89984",
			Accent:           "#D3869B",
			Success:          "#B8BB26",
			Warning:          "#FABD2F",
			Error:            "#FB4934",
			Info:             "#83A598",
			Border:           "#3C3836",
			BorderFocused:    "#83A598",
			Selection:        "#504945",
			Muted:            "#928374",
			UserMessage:      "#83A598",
			AssistantMessage: "#B8BB26",
			SystemMessage:    "#FABD2F",
			Keyword:          "#FB4934",
			String:           "#B8BB26",
			Number:           "#D3869B",
			Comment:          "#928374",
			Function:         "#8EC07C",
		},
	}
}

// OneDark returns the One Dark theme
func OneDark() *Theme {
	return &Theme{
		Name:        "one-dark",
		Author:      "Atom",
		Description: "Atom's iconic One Dark theme",
		Colors: ThemeColors{
			Background:       "#282C34",
			Foreground:       "#ABB2BF",
			Primary:          "#61AFEF",
			Secondary:        "#5C6370",
			Accent:           "#C678DD",
			Success:          "#98C379",
			Warning:          "#E5C07B",
			Error:            "#E06C75",
			Info:             "#56B6C2",
			Border:           "#3E4451",
			BorderFocused:    "#61AFEF",
			Selection:        "#3E4451",
			Muted:            "#5C6370",
			UserMessage:      "#61AFEF",
			AssistantMessage: "#98C379",
			SystemMessage:    "#E5C07B",
			Keyword:          "#C678DD",
			String:           "#98C379",
			Number:           "#D19A66",
			Comment:          "#5C6370",
			Function:         "#61AFEF",
		},
	}
}

// SolarizedDark returns the Solarized Dark theme
func SolarizedDark() *Theme {
	return &Theme{
		Name:        "solarized-dark",
		Author:      "Ethan Schoonover",
		Description: "Precision colors for machines and people",
		Colors: ThemeColors{
			Background:       "#002B36",
			Foreground:       "#839496",
			Primary:          "#268BD2",
			Secondary:        "#586E75",
			Accent:           "#D33682",
			Success:          "#859900",
			Warning:          "#B58900",
			Error:            "#DC322F",
			Info:             "#2AA198",
			Border:           "#073642",
			BorderFocused:    "#268BD2",
			Selection:        "#073642",
			Muted:            "#657B83",
			UserMessage:      "#268BD2",
			AssistantMessage: "#859900",
			SystemMessage:    "#B58900",
			Keyword:          "#859900",
			String:           "#2AA198",
			Number:           "#D33682",
			Comment:          "#586E75",
			Function:         "#268BD2",
		},
	}
}

// Monokai returns the Monokai theme
func Monokai() *Theme {
	return &Theme{
		Name:        "monokai",
		Author:      "Wimer Hazenberg",
		Description: "The famous Monokai color scheme",
		Colors: ThemeColors{
			Background:       "#272822",
			Foreground:       "#F8F8F2",
			Primary:          "#66D9EF",
			Secondary:        "#75715E",
			Accent:           "#F92672",
			Success:          "#A6E22E",
			Warning:          "#E6DB74",
			Error:            "#F92672",
			Info:             "#66D9EF",
			Border:           "#3E3D32",
			BorderFocused:    "#66D9EF",
			Selection:        "#49483E",
			Muted:            "#75715E",
			UserMessage:      "#66D9EF",
			AssistantMessage: "#A6E22E",
			SystemMessage:    "#E6DB74",
			Keyword:          "#F92672",
			String:           "#E6DB74",
			Number:           "#AE81FF",
			Comment:          "#75715E",
			Function:         "#A6E22E",
		},
	}
}

// GitHubDark returns the GitHub Dark theme
func GitHubDark() *Theme {
	return &Theme{
		Name:        "github-dark",
		Author:      "GitHub",
		Description: "GitHub's official dark theme",
		Colors: ThemeColors{
			Background:       "#0D1117",
			Foreground:       "#C9D1D9",
			Primary:          "#58A6FF",
			Secondary:        "#8B949E",
			Accent:           "#F778BA",
			Success:          "#3FB950",
			Warning:          "#D29922",
			Error:            "#F85149",
			Info:             "#58A6FF",
			Border:           "#30363D",
			BorderFocused:    "#58A6FF",
			Selection:        "#161B22",
			Muted:            "#484F58",
			UserMessage:      "#58A6FF",
			AssistantMessage: "#3FB950",
			SystemMessage:    "#D29922",
			Keyword:          "#FF7B72",
			String:           "#A5D6FF",
			Number:           "#79C0FF",
			Comment:          "#8B949E",
			Function:         "#D2A8FF",
		},
	}
}
