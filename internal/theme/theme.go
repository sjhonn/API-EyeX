package theme

import "github.com/eyex-api/eyex/internal/models"

var supportedTypes = []string{
	"normal",
	"protanopia",
	"deuteranopia",
	"tritanopia",
	"achromatopsia",
}

var palettes = map[string]models.Palette{
	"normal": {
		Background: "#F4F5F7",
		Surface:    "#FFFFFF",
		Text:       "#20252B",
		Primary:    "#2E6DA4",
		Secondary:  "#6B7785",
		Error:      "#C94C4C",
		Success:    "#3C8D5A",
	},
	"protanopia": {
		Background: "#1E1E1E",
		Surface:    "#2A2A2A",
		Text:       "#F5F5F5",
		Primary:    "#3F8FD2",
		Secondary:  "#E3B341",
		Error:      "#D96C3F",
		Success:    "#4FB3A5",
	},
	"deuteranopia": {
		Background: "#1E1E1E",
		Surface:    "#2A2A2A",
		Text:       "#F5F5F5",
		Primary:    "#4A90D9",
		Secondary:  "#D9A24A",
		Error:      "#D94A4A",
		Success:    "#4AD98C",
	},
	"tritanopia": {
		Background: "#202124",
		Surface:    "#2D2F33",
		Text:       "#F5F5F5",
		Primary:    "#D65DB1",
		Secondary:  "#4CC9A7",
		Error:      "#E05A47",
		Success:    "#64A66F",
	},
	"achromatopsia": {
		Background: "#202020",
		Surface:    "#303030",
		Text:       "#F2F2F2",
		Primary:    "#D0D0D0",
		Secondary:  "#A8A8A8",
		Error:      "#E0E0E0",
		Success:    "#BEBEBE",
	},
}

func Types() []string {
	out := make([]string, len(supportedTypes))
	copy(out, supportedTypes)
	return out
}

func Get(typeValue string) (models.ThemeResponse, bool) {
	palette, ok := palettes[typeValue]
	if !ok {
		return models.ThemeResponse{}, false
	}
	return models.ThemeResponse{Type: typeValue, Palette: palette}, true
}
