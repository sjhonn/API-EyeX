package colorblind

import "testing"

func TestNormalizeHex(t *testing.T) {
	got, err := NormalizeHex("#33aa99")
	if err != nil {
		t.Fatal(err)
	}
	if got != "#33AA99" {
		t.Fatalf("expected #33AA99, got %s", got)
	}
}

func TestInvalidHex(t *testing.T) {
	if _, err := NormalizeHex("336699"); err == nil {
		t.Fatal("expected invalid hex error")
	}
}

func TestSimulationIsDeterministic(t *testing.T) {
	first, err := SimulateHex("#FF0000", Protanopia)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SimulateHex("#FF0000", Protanopia)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("simulation must be deterministic: %s != %s", first, second)
	}
}

func TestAcromatopsiaProducesGray(t *testing.T) {
	got, err := SimulateHex("#3A7BD5", Acromatopsia)
	if err != nil {
		t.Fatal(err)
	}
	rgb, err := ParseHex(got)
	if err != nil {
		t.Fatal(err)
	}
	if rgb.R != rgb.G || rgb.G != rgb.B {
		t.Fatalf("expected grayscale, got %s", got)
	}
}

func TestPaletteCount(t *testing.T) {
	got, err := PaletteHex("#336699", Deuteranopia, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(got))
	}
}

func TestMachadoContinuousSeverity(t *testing.T) {
	got, err := SimulateHexSeverity("#FF0000", Protanopia, 0.65)
	if err != nil {
		t.Fatal(err)
	}
	if got != "#A05A00" {
		t.Fatalf("expected #A05A00, got %s", got)
	}
}

func TestMachadoZeroSeverityIsIdentity(t *testing.T) {
	got, err := SimulateHexSeverity("#12abEF", Deuteranopia, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != "#12ABEF" {
		t.Fatalf("expected identity color, got %s", got)
	}
}

func TestMachadoRejectsSeverityOutsideRange(t *testing.T) {
	if _, err := SimulateHexSeverity("#336699", Tritanopia, -0.1); err == nil {
		t.Fatal("expected severity error")
	}
	if _, err := SimulateHexSeverity("#336699", Tritanopia, 1.1); err == nil {
		t.Fatal("expected severity error")
	}
}
