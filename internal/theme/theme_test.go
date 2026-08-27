package theme

import "testing"

func TestGeneratedThemesMeetTextContrast(t *testing.T) {
	for _, typeValue := range Types() {
		for _, mode := range []string{ModeLight, ModeDark} {
			for _, severity := range []string{SeverityMild, SeverityModerate, SeveritySevere} {
				for _, highContrast := range []bool{false, true} {
					result, ok, err := Get(typeValue, Options{Severity: severity, Mode: mode, HighContrast: highContrast, Explicit: true})
					if err != nil || !ok {
						t.Fatalf("%s/%s/%s/%v failed: ok=%v err=%v", typeValue, mode, severity, highContrast, ok, err)
					}
					if !result.ContrastOK {
						t.Fatalf("%s/%s/%s/%v did not meet WCAG text contrast: %#v", typeValue, mode, severity, highContrast, result.Palette)
					}
				}
			}
		}
	}
}

func TestLegacyPalettesRemainUnchanged(t *testing.T) {
	for typeValue, expected := range legacyPalettes {
		result, ok, err := Get(typeValue, Options{})
		if err != nil || !ok {
			t.Fatalf("legacy %s failed: %v", typeValue, err)
		}
		if result.Palette != expected {
			t.Fatalf("legacy palette changed for %s: %#v", typeValue, result.Palette)
		}
	}
}
