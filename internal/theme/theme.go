package theme

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/eyex-api/eyex/internal/models"
)

const (
	SeverityMild     = "mild"
	SeverityModerate = "moderate"
	SeveritySevere   = "severe"
	ModeDark         = "dark"
	ModeLight        = "light"
)

var supportedTypes = []string{
	"normal",
	"protanopia",
	"deuteranopia",
	"tritanopia",
	"achromatopsia",
	"low_vision",
}

// legacyPalettes are intentionally unchanged so existing clients get the same
// colors when no new query parameter is supplied.
var legacyPalettes = map[string]models.Palette{
	"normal": {
		Background: "#F4F5F7", Surface: "#FFFFFF", Text: "#20252B", Primary: "#2E6DA4",
		Secondary: "#6B7785", Error: "#C94C4C", Success: "#3C8D5A",
	},
	"protanopia": {
		Background: "#1E1E1E", Surface: "#2A2A2A", Text: "#F5F5F5", Primary: "#3F8FD2",
		Secondary: "#E3B341", Error: "#D96C3F", Success: "#4FB3A5",
	},
	"deuteranopia": {
		Background: "#1E1E1E", Surface: "#2A2A2A", Text: "#F5F5F5", Primary: "#4A90D9",
		Secondary: "#D9A24A", Error: "#D94A4A", Success: "#4AD98C",
	},
	"tritanopia": {
		Background: "#202124", Surface: "#2D2F33", Text: "#F5F5F5", Primary: "#D65DB1",
		Secondary: "#4CC9A7", Error: "#E05A47", Success: "#64A66F",
	},
	"achromatopsia": {
		Background: "#202020", Surface: "#303030", Text: "#F2F2F2", Primary: "#D0D0D0",
		Secondary: "#A8A8A8", Error: "#E0E0E0", Success: "#BEBEBE",
	},
}

var safePalettes = map[string]map[string]models.Palette{
	"normal": {
		ModeLight: legacyPalettes["normal"],
		ModeDark:  {Background: "#181A1D", Surface: "#24272B", Text: "#F5F7FA", Primary: "#5CA9E6", Secondary: "#AAB4BE", Error: "#FF7B72", Success: "#56D364"},
	},
	"protanopia": {
		ModeLight: {Background: "#F7F8FA", Surface: "#FFFFFF", Text: "#1D2329", Primary: "#256EA6", Secondary: "#916B00", Error: "#A84824", Success: "#237A70"},
		ModeDark:  legacyPalettes["protanopia"],
	},
	"deuteranopia": {
		ModeLight: {Background: "#F7F8FA", Surface: "#FFFFFF", Text: "#1D2329", Primary: "#236FAE", Secondary: "#8A6200", Error: "#A83D3D", Success: "#187A55"},
		ModeDark:  legacyPalettes["deuteranopia"],
	},
	"tritanopia": {
		ModeLight: {Background: "#F7F7F8", Surface: "#FFFFFF", Text: "#202124", Primary: "#9B3F80", Secondary: "#167A65", Error: "#AA4234", Success: "#347A42"},
		ModeDark:  legacyPalettes["tritanopia"],
	},
	"achromatopsia": {
		ModeLight: {Background: "#FAFAFA", Surface: "#FFFFFF", Text: "#181818", Primary: "#4A4A4A", Secondary: "#666666", Error: "#303030", Success: "#555555"},
		ModeDark:  legacyPalettes["achromatopsia"],
	},
	"low_vision": {
		ModeLight: {Background: "#FFFFFF", Surface: "#F2F2F2", Text: "#000000", Primary: "#005FCC", Secondary: "#6D4C00", Error: "#A80000", Success: "#006B35"},
		ModeDark:  {Background: "#000000", Surface: "#121212", Text: "#FFFFFF", Primary: "#66B2FF", Secondary: "#FFD166", Error: "#FF6B6B", Success: "#65E6A3"},
	},
}

var hexPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type Options struct {
	Severity     string
	Mode         string
	HighContrast bool
	Explicit     bool
}

func Types() []string {
	out := make([]string, len(supportedTypes))
	copy(out, supportedTypes)
	return out
}

func IsSupportedType(value string) bool {
	for _, item := range supportedTypes {
		if value == item {
			return true
		}
	}
	return false
}

func ValidateOptions(options Options) error {
	if options.Severity != "" && options.Severity != SeverityMild && options.Severity != SeverityModerate && options.Severity != SeveritySevere {
		return errors.New("severity debe ser mild, moderate o severe")
	}
	if options.Mode != "" && options.Mode != ModeDark && options.Mode != ModeLight {
		return errors.New("mode debe ser dark o light")
	}
	return nil
}

func Get(typeValue string, options Options) (models.ThemeResponse, bool, error) {
	if !IsSupportedType(typeValue) {
		return models.ThemeResponse{}, false, nil
	}
	if err := ValidateOptions(options); err != nil {
		return models.ThemeResponse{}, true, err
	}

	if !options.Explicit && typeValue != "low_vision" {
		palette := legacyPalettes[typeValue]
		return response(typeValue, palette), true, nil
	}

	mode := options.Mode
	if mode == "" {
		if typeValue == "normal" {
			mode = ModeLight
		} else {
			mode = ModeDark
		}
	}
	severity := options.Severity
	if severity == "" {
		severity = SeverityModerate
	}

	palette := generatedPalette(typeValue, mode, severity)
	if options.HighContrast || typeValue == "low_vision" {
		palette = applyHighContrast(palette, typeValue, mode)
	}
	palette = ensureTextContrast(palette)
	return response(typeValue, palette), true, nil
}

func Custom(req models.CustomThemeRequest) (models.ThemeResponse, error) {
	if !IsSupportedType(req.Type) {
		return models.ThemeResponse{}, errors.New("invalid_type")
	}
	if err := ValidatePalette(req.Palette); err != nil {
		return models.ThemeResponse{}, err
	}
	options := Options{Severity: req.Severity, Mode: req.Mode, HighContrast: req.HighContrast, Explicit: true}
	if err := ValidateOptions(options); err != nil {
		return models.ThemeResponse{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = inferMode(req.Palette.Background)
	}
	severity := req.Severity
	if severity == "" {
		severity = SeverityModerate
	}

	palette := req.Palette
	factor := severityFactor(severity)
	if req.Type == "achromatopsia" {
		palette = models.Palette{
			Background: mixHex(palette.Background, grayscaleHex(palette.Background), factor),
			Surface:    mixHex(palette.Surface, grayscaleHex(palette.Surface), factor),
			Text:       mixHex(palette.Text, grayscaleHex(palette.Text), factor),
			Primary:    mixHex(palette.Primary, grayscaleHex(palette.Primary), factor),
			Secondary:  mixHex(palette.Secondary, grayscaleHex(palette.Secondary), factor),
			Error:      mixHex(palette.Error, grayscaleHex(palette.Error), factor),
			Success:    mixHex(palette.Success, grayscaleHex(palette.Success), factor),
		}
	} else if req.Type != "normal" {
		anchor := safePalettes[req.Type][mode]
		palette.Primary = mixHex(palette.Primary, anchor.Primary, factor)
		palette.Secondary = mixHex(palette.Secondary, anchor.Secondary, factor)
		palette.Error = mixHex(palette.Error, anchor.Error, factor)
		palette.Success = mixHex(palette.Success, anchor.Success, factor)
	}
	palette = adaptMode(palette, mode)
	if req.HighContrast || req.Type == "low_vision" {
		palette = applyHighContrast(palette, req.Type, mode)
	}
	palette = ensureTextContrast(palette)
	return response(req.Type, palette), nil
}

func ValidatePalette(p models.Palette) error {
	values := map[string]string{
		"background": p.Background, "surface": p.Surface, "text": p.Text,
		"primary": p.Primary, "secondary": p.Secondary, "error": p.Error, "success": p.Success,
	}
	for name, value := range values {
		if !hexPattern.MatchString(value) {
			return fmt.Errorf("%s debe usar formato #RRGGBB", name)
		}
	}
	return nil
}

func ContrastOK(p models.Palette) bool {
	return ContrastRatio(p.Text, p.Background) >= 4.5 && ContrastRatio(p.Text, p.Surface) >= 4.5
}

func ContrastRatio(a, b string) float64 {
	la := relativeLuminance(a)
	lb := relativeLuminance(b)
	lighter, darker := math.Max(la, lb), math.Min(la, lb)
	return (lighter + 0.05) / (darker + 0.05)
}

func SuggestType(answers models.QuickTestAnswers) string {
	if answers.ColorsLookGray {
		return "achromatopsia"
	}
	if answers.BlueYellowConfusion {
		return "tritanopia"
	}
	if answers.RedsLookDarker && answers.GreenBrownConfusion {
		return "protanopia"
	}
	if answers.GreenBrownConfusion {
		return "deuteranopia"
	}
	if answers.RedsLookDarker {
		return "protanopia"
	}
	return "normal"
}

func generatedPalette(typeValue, mode, severity string) models.Palette {
	anchor := safePalettes[typeValue][mode]
	if typeValue == "normal" {
		return anchor
	}
	normal := safePalettes["normal"][mode]
	return mixPalette(normal, anchor, severityFactor(severity))
}

func severityFactor(severity string) float64 {
	switch severity {
	case SeverityMild:
		return 0.35
	case SeveritySevere:
		return 1
	default:
		return 0.70
	}
}

func mixPalette(a, b models.Palette, factor float64) models.Palette {
	return models.Palette{
		Background: mixHex(a.Background, b.Background, factor),
		Surface:    mixHex(a.Surface, b.Surface, factor),
		Text:       mixHex(a.Text, b.Text, factor),
		Primary:    mixHex(a.Primary, b.Primary, factor),
		Secondary:  mixHex(a.Secondary, b.Secondary, factor),
		Error:      mixHex(a.Error, b.Error, factor),
		Success:    mixHex(a.Success, b.Success, factor),
	}
}

func adaptMode(p models.Palette, mode string) models.Palette {
	if mode == ModeDark {
		p.Background = mixHex(p.Background, "#181A1D", 0.72)
		p.Surface = mixHex(p.Surface, "#24272B", 0.72)
		if ContrastRatio(p.Text, p.Background) < 4.5 || ContrastRatio(p.Text, p.Surface) < 4.5 {
			p.Text = "#F5F7FA"
		}
		return p
	}
	p.Background = mixHex(p.Background, "#F4F6F8", 0.72)
	p.Surface = mixHex(p.Surface, "#FFFFFF", 0.82)
	if ContrastRatio(p.Text, p.Background) < 4.5 || ContrastRatio(p.Text, p.Surface) < 4.5 {
		p.Text = "#1A1D21"
	}
	return p
}

func applyHighContrast(p models.Palette, typeValue, mode string) models.Palette {
	anchorType := typeValue
	if anchorType == "normal" {
		anchorType = "low_vision"
	}
	anchor, ok := safePalettes[anchorType][mode]
	if !ok {
		anchor = safePalettes["low_vision"][mode]
	}
	if mode == ModeDark {
		p.Background, p.Surface, p.Text = "#000000", "#121212", "#FFFFFF"
	} else {
		p.Background, p.Surface, p.Text = "#FFFFFF", "#F2F2F2", "#000000"
	}
	p.Primary = mixHex(p.Primary, anchor.Primary, 0.45)
	p.Secondary = mixHex(p.Secondary, anchor.Secondary, 0.45)
	p.Error = mixHex(p.Error, anchor.Error, 0.45)
	p.Success = mixHex(p.Success, anchor.Success, 0.45)
	return p
}

func ensureTextContrast(p models.Palette) models.Palette {
	if ContrastOK(p) {
		return p
	}
	whiteScore := math.Min(ContrastRatio("#FFFFFF", p.Background), ContrastRatio("#FFFFFF", p.Surface))
	blackScore := math.Min(ContrastRatio("#000000", p.Background), ContrastRatio("#000000", p.Surface))
	if whiteScore >= blackScore {
		p.Text = "#FFFFFF"
	} else {
		p.Text = "#000000"
	}
	return p
}

func inferMode(background string) string {
	if relativeLuminance(background) < 0.35 {
		return ModeDark
	}
	return ModeLight
}

func response(typeValue string, p models.Palette) models.ThemeResponse {
	return models.ThemeResponse{Type: typeValue, Palette: p, ContrastOK: ContrastOK(p)}
}

func grayscaleHex(value string) string {
	r, g, b := parseHex(value)
	gray := clampInt(int(math.Round(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))))
	return formatHex(gray, gray, gray)
}

func mixHex(a, b string, factor float64) string {
	ar, ag, ab := parseHex(a)
	br, bg, bb := parseHex(b)
	mix := func(x, y int) int { return clampInt(int(math.Round(float64(x)*(1-factor) + float64(y)*factor))) }
	return formatHex(mix(ar, br), mix(ag, bg), mix(ab, bb))
}

func relativeLuminance(hex string) float64 {
	r, g, b := parseHex(hex)
	linear := func(v int) float64 {
		x := float64(v) / 255
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(r) + 0.7152*linear(g) + 0.0722*linear(b)
}

func parseHex(value string) (int, int, int) {
	value = strings.TrimPrefix(value, "#")
	if len(value) != 6 {
		return 0, 0, 0
	}
	parse := func(s string) int {
		v, err := strconv.ParseInt(s, 16, 0)
		if err != nil {
			return 0
		}
		return int(v)
	}
	return parse(value[0:2]), parse(value[2:4]), parse(value[4:6])
}

func formatHex(r, g, b int) string {
	return fmt.Sprintf("#%02X%02X%02X", clampInt(r), clampInt(g), clampInt(b))
}

func clampInt(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}
