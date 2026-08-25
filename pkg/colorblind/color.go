package colorblind

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type RGB struct {
	R float64
	G float64
	B float64
}

type Deficiency string

const (
	Protanopia    Deficiency = "protanopia"
	Protanomaly   Deficiency = "protanomaly"
	Deuteranopia  Deficiency = "deuteranopia"
	Deuteranomaly Deficiency = "deuteranomaly"
	Tritanopia    Deficiency = "tritanopia"
	Tritanomaly   Deficiency = "tritanomaly"
	Acromatopsia  Deficiency = "acromatopsia"
)

type TypeInfo struct {
	ID          string
	Name        string
	Description string
}

var ErrInvalidHex = errors.New("invalid hex color")
var ErrInvalidType = errors.New("unsupported color vision deficiency type")

var supportedTypes = []TypeInfo{
	{string(Protanopia), "Protanopia", "Ausencia funcional de conos sensibles al rojo."},
	{string(Protanomaly), "Protanomalía", "Sensibilidad reducida al rojo."},
	{string(Deuteranopia), "Deuteranopia", "Ausencia funcional de conos sensibles al verde."},
	{string(Deuteranomaly), "Deuteranomalía", "Sensibilidad reducida al verde."},
	{string(Tritanopia), "Tritanopia", "Ausencia funcional de conos sensibles al azul."},
	{string(Tritanomaly), "Tritanomalía", "Sensibilidad reducida al azul."},
	{string(Acromatopsia), "Acromatopsia", "Visión monocromática; el color se reduce a luminancia."},
}

func Types() []TypeInfo {
	out := make([]TypeInfo, len(supportedTypes))
	copy(out, supportedTypes)
	return out
}

func ParseDeficiency(value string) (Deficiency, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "protanopia":
		return Protanopia, nil
	case "protanomaly", "protanomalia", "protanomalía":
		return Protanomaly, nil
	case "deuteranopia":
		return Deuteranopia, nil
	case "deuteranomaly", "deuteranomalia", "deuteranomalía":
		return Deuteranomaly, nil
	case "tritanopia":
		return Tritanopia, nil
	case "tritanomaly", "tritanomalia", "tritanomalía":
		return Tritanomaly, nil
	case "acromatopsia", "achromatopsia":
		return Acromatopsia, nil
	default:
		return "", ErrInvalidType
	}
}

func ParseHex(value string) (RGB, error) {
	value = strings.TrimSpace(value)
	if len(value) != 7 || value[0] != '#' {
		return RGB{}, ErrInvalidHex
	}

	n, err := strconv.ParseUint(value[1:], 16, 24)
	if err != nil {
		return RGB{}, ErrInvalidHex
	}

	return RGB{
		R: float64((n>>16)&0xff) / 255,
		G: float64((n>>8)&0xff) / 255,
		B: float64(n&0xff) / 255,
	}, nil
}

func Hex(rgb RGB) string {
	r := int(math.Round(clamp01(rgb.R) * 255))
	g := int(math.Round(clamp01(rgb.G) * 255))
	b := int(math.Round(clamp01(rgb.B) * 255))
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func NormalizeHex(value string) (string, error) {
	rgb, err := ParseHex(value)
	if err != nil {
		return "", err
	}
	return Hex(rgb), nil
}

func SimulateHex(hex string, deficiency Deficiency) (string, error) {
	rgb, err := ParseHex(hex)
	if err != nil {
		return "", err
	}
	simulated := Simulate(rgb, deficiency)
	return Hex(simulated), nil
}

func Simulate(rgb RGB, deficiency Deficiency) RGB {
	linear := RGB{srgbToLinear(rgb.R), srgbToLinear(rgb.G), srgbToLinear(rgb.B)}

	if deficiency == Acromatopsia {
		y := 0.2126*linear.R + 0.7152*linear.G + 0.0722*linear.B
		return RGB{linearToSRGB(y), linearToSRGB(y), linearToSRGB(y)}
	}

	matrix, severity := matrixFor(deficiency)
	transformed := applyMatrix(linear, matrix)
	if severity < 1 {
		transformed = RGB{
			R: linear.R + (transformed.R-linear.R)*severity,
			G: linear.G + (transformed.G-linear.G)*severity,
			B: linear.B + (transformed.B-linear.B)*severity,
		}
	}

	return RGB{
		R: linearToSRGB(clamp01(transformed.R)),
		G: linearToSRGB(clamp01(transformed.G)),
		B: linearToSRGB(clamp01(transformed.B)),
	}
}

func CorrectHex(hex string, deficiency Deficiency) (corrected string, diagnosticContrast float64, err error) {
	rgb, err := ParseHex(hex)
	if err != nil {
		return "", 0, err
	}

	// With total monochromacy there is no chromatic channel to compensate.
	// Returning the luminance-equivalent gray is explicit and predictable.
	if deficiency == Acromatopsia {
		gray := Simulate(rgb, deficiency)
		return Hex(gray), round2(contrastRatio(rgb, gray)), nil
	}

	originalSim := Simulate(rgb, deficiency)
	targetOK := toOKLab(rgb)
	baseError := okDistance(targetOK, toOKLab(originalSim))
	originalLum := relativeLuminance(rgb)

	best := rgb
	bestScore := 0.0 // Keeping the original is always a valid fallback.

	for _, candidate := range correctionCandidates(rgb) {
		simulated := Simulate(candidate, deficiency)
		fidelityError := okDistance(targetOK, toOKLab(simulated))
		improvement := baseError - fidelityError
		normalChange := okDistance(targetOK, toOKLab(candidate))
		lumDelta := math.Abs(relativeLuminance(candidate) - originalLum)

		// The primary goal is compensatory: after simulation, the candidate should
		// move perceptually closer to the originally intended color. Penalize
		// unnecessarily large changes and luminance jumps.
		score := (5.0 * improvement) - (0.85 * normalChange) - (0.35 * lumDelta)
		if improvement < 0.004 {
			score -= 0.1
		}

		if score > bestScore {
			bestScore = score
			best = candidate
		}
	}

	correctedSim := Simulate(best, deficiency)
	ratio := contrastRatio(originalSim, correctedSim)
	return Hex(best), round2(ratio), nil
}

func PaletteHex(hex string, deficiency Deficiency, count int) ([]string, error) {
	rgb, err := ParseHex(hex)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		return []string{}, nil
	}

	if deficiency == Acromatopsia {
		return achromatopsiaPalette(rgb, count), nil
	}

	originalSim := Simulate(rgb, deficiency)
	originalOK := toOKLab(rgb)
	originalLum := relativeLuminance(rgb)

	type scored struct {
		rgb   RGB
		score float64
	}
	pool := make([]scored, 0, 128)
	seen := map[string]struct{}{Hex(rgb): {}}

	for _, candidate := range correctionCandidates(rgb) {
		key := Hex(candidate)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		separation := okDistance(toOKLab(originalSim), toOKLab(Simulate(candidate, deficiency)))
		normalChange := okDistance(originalOK, toOKLab(candidate))
		lumDelta := math.Abs(relativeLuminance(candidate) - originalLum)
		score := (3.8 * separation) - (1.15 * normalChange) - (0.5 * lumDelta)
		pool = append(pool, scored{candidate, score})
	}

	// Small pool: selection sort keeps the package dependency-free and clear.
	for i := 0; i < len(pool); i++ {
		best := i
		for j := i + 1; j < len(pool); j++ {
			if pool[j].score > pool[best].score {
				best = j
			}
		}
		pool[i], pool[best] = pool[best], pool[i]
	}

	out := make([]string, 0, count)
	for _, item := range pool {
		candidateHex := Hex(item.rgb)
		diverse := true
		for _, existing := range out {
			existingRGB, _ := ParseHex(existing)
			if okDistance(toOKLab(existingRGB), toOKLab(item.rgb)) < 0.055 {
				diverse = false
				break
			}
		}
		if diverse {
			out = append(out, candidateHex)
		}
		if len(out) == count {
			break
		}
	}

	return out, nil
}

func matrixFor(d Deficiency) ([3][3]float64, float64) {
	identity := [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	switch d {
	case Protanopia, Protanomaly:
		return [3][3]float64{
			{0.152286, 1.052583, -0.204868},
			{0.114503, 0.786281, 0.099216},
			{-0.003882, -0.048116, 1.051998},
		}, anomalySeverity(d)
	case Deuteranopia, Deuteranomaly:
		return [3][3]float64{
			{0.367322, 0.860646, -0.227968},
			{0.280085, 0.672501, 0.047413},
			{-0.011820, 0.042940, 0.968881},
		}, anomalySeverity(d)
	case Tritanopia, Tritanomaly:
		return [3][3]float64{
			{1.255528, -0.076749, -0.178779},
			{-0.078411, 0.930809, 0.147602},
			{0.004733, 0.691367, 0.303900},
		}, anomalySeverity(d)
	default:
		return identity, 1
	}
}

func anomalySeverity(d Deficiency) float64 {
	switch d {
	case Protanomaly, Deuteranomaly, Tritanomaly:
		return 0.6
	default:
		return 1
	}
}

func applyMatrix(rgb RGB, m [3][3]float64) RGB {
	return RGB{
		R: m[0][0]*rgb.R + m[0][1]*rgb.G + m[0][2]*rgb.B,
		G: m[1][0]*rgb.R + m[1][1]*rgb.G + m[1][2]*rgb.B,
		B: m[2][0]*rgb.R + m[2][1]*rgb.G + m[2][2]*rgb.B,
	}
}

func achromatopsiaPalette(rgb RGB, count int) []string {
	_, _, lightness := rgbToHSL(Simulate(rgb, Acromatopsia))
	offsets := []float64{-0.28, 0.28, -0.16, 0.16, -0.08, 0.08}
	out := make([]string, 0, count)
	seen := make(map[string]struct{})
	for _, offset := range offsets {
		gray := hslToRGB(0, 0, clamp(lightness+offset, 0.08, 0.92))
		hex := Hex(gray)
		if _, exists := seen[hex]; exists {
			continue
		}
		seen[hex] = struct{}{}
		out = append(out, hex)
		if len(out) == count {
			break
		}
	}
	return out
}

func correctionCandidates(rgb RGB) []RGB {
	h, s, l := rgbToHSL(rgb)
	hueOffsets := []float64{-160, -120, -90, -60, -35, -20, 20, 35, 60, 90, 120, 160, 180}
	saturationOffsets := []float64{-0.12, 0, 0.10, 0.20}
	lightnessOffsets := []float64{-0.18, -0.10, -0.05, 0, 0.05, 0.10, 0.18}

	candidates := make([]RGB, 0, len(hueOffsets)*len(saturationOffsets)*len(lightnessOffsets)+8)
	for _, dh := range hueOffsets {
		for _, ds := range saturationOffsets {
			for _, dl := range lightnessOffsets {
				ns := clamp01(s + ds)
				nl := clamp(l+dl, 0.08, 0.92)
				candidates = append(candidates, hslToRGB(wrapHue(h+dh), ns, nl))
			}
		}
	}

	// Include small luminance-only alternatives for already distinguishable colors.
	for _, dl := range []float64{-0.22, -0.14, -0.07, 0.07, 0.14, 0.22} {
		candidates = append(candidates, hslToRGB(h, s, clamp(l+dl, 0.08, 0.92)))
	}
	return candidates
}

func rgbToHSL(rgb RGB) (h, s, l float64) {
	maxv := math.Max(rgb.R, math.Max(rgb.G, rgb.B))
	minv := math.Min(rgb.R, math.Min(rgb.G, rgb.B))
	l = (maxv + minv) / 2
	if maxv == minv {
		return 0, 0, l
	}

	d := maxv - minv
	if l > 0.5 {
		s = d / (2 - maxv - minv)
	} else {
		s = d / (maxv + minv)
	}

	switch maxv {
	case rgb.R:
		h = (rgb.G - rgb.B) / d
		if rgb.G < rgb.B {
			h += 6
		}
	case rgb.G:
		h = (rgb.B-rgb.R)/d + 2
	default:
		h = (rgb.R-rgb.G)/d + 4
	}
	h *= 60
	return
}

func hslToRGB(h, s, l float64) RGB {
	h = wrapHue(h) / 360
	if s == 0 {
		return RGB{l, l, l}
	}

	q := l * (1 + s)
	if l >= 0.5 {
		q = l + s - l*s
	}
	p := 2*l - q

	return RGB{
		R: hueToRGB(p, q, h+1.0/3.0),
		G: hueToRGB(p, q, h),
		B: hueToRGB(p, q, h-1.0/3.0),
	}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}

type okLab struct{ L, A, B float64 }

func toOKLab(rgb RGB) okLab {
	r := srgbToLinear(rgb.R)
	g := srgbToLinear(rgb.G)
	b := srgbToLinear(rgb.B)

	l := 0.4122214708*r + 0.5363325363*g + 0.0514459929*b
	m := 0.2119034982*r + 0.6806995451*g + 0.1073969566*b
	s := 0.0883024619*r + 0.2817188376*g + 0.6299787005*b

	l = math.Cbrt(l)
	m = math.Cbrt(m)
	s = math.Cbrt(s)

	return okLab{
		L: 0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		A: 1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		B: 0.0259040371*l + 0.7827717662*m - 0.8086757660*s,
	}
}

func okDistance(a, b okLab) float64 {
	dl, da, db := a.L-b.L, a.A-b.A, a.B-b.B
	return math.Sqrt(dl*dl + da*da + db*db)
}

func contrastRatio(a, b RGB) float64 {
	l1 := relativeLuminance(a)
	l2 := relativeLuminance(b)
	if l2 > l1 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(rgb RGB) float64 {
	return 0.2126*srgbToLinear(rgb.R) + 0.7152*srgbToLinear(rgb.G) + 0.0722*srgbToLinear(rgb.B)
}

func srgbToLinear(v float64) float64 {
	v = clamp01(v)
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func linearToSRGB(v float64) float64 {
	v = clamp01(v)
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}

func wrapHue(v float64) float64 {
	v = math.Mod(v, 360)
	if v < 0 {
		v += 360
	}
	return v
}

func clamp01(v float64) float64 { return clamp(v, 0, 1) }
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
