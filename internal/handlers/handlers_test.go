package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/eyex-api/eyex/internal/models"
)

func TestThemeTypes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/types", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var payload models.TypesResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	want := []string{"normal", "protanopia", "deuteranopia", "tritanopia", "achromatopsia", "low_vision"}
	if !reflect.DeepEqual(payload.Types, want) {
		t.Fatalf("unexpected types: %#v", payload.Types)
	}
}

func TestBaseContractKeepsLegacyPalette(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/deuteranopia", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	var payload models.ThemeResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if payload.Type != "deuteranopia" {
		t.Fatalf("unexpected type: %q", payload.Type)
	}
	if payload.Palette.Background != "#1E1E1E" || payload.Palette.Surface != "#2A2A2A" || payload.Palette.Text != "#F5F5F5" || payload.Palette.Primary != "#4A90D9" || payload.Palette.Secondary != "#D9A24A" || payload.Palette.Error != "#D94A4A" || payload.Palette.Success != "#4AD98C" {
		t.Fatalf("legacy palette changed: %#v", payload.Palette)
	}
	if !payload.ContrastOK {
		t.Fatalf("expected legacy palette to pass WCAG text contrast")
	}
}

func TestQueryOptionsGenerateAccessibleTheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/protanopia?severity=mild&mode=light&high_contrast=true", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.ThemeResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if !payload.ContrastOK {
		t.Fatalf("expected contrast_ok=true: %#v", payload)
	}
	if payload.Palette.Background != "#FFFFFF" || payload.Palette.Text != "#000000" {
		t.Fatalf("high contrast light theme not applied: %#v", payload.Palette)
	}
}

func TestLowVisionTheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/low_vision?mode=dark", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.ThemeResponse
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if !payload.ContrastOK || payload.Palette.Background != "#000000" || payload.Palette.Text != "#FFFFFF" {
		t.Fatalf("unexpected low vision palette: %#v", payload)
	}
}

func TestInvalidQueryParameter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/protanopia?severity=extreme", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}
}

func TestCustomTheme(t *testing.T) {
	body := []byte(`{
		"type":"deuteranopia",
		"severity":"moderate",
		"mode":"dark",
		"high_contrast":true,
		"palette":{
			"background":"#101820",
			"surface":"#182430",
			"text":"#F8F9FA",
			"primary":"#E63946",
			"secondary":"#2A9D8F",
			"error":"#D62828",
			"success":"#2A9D8F"
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/theme/custom", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.ThemeResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if payload.Type != "deuteranopia" || !payload.ContrastOK {
		t.Fatalf("unexpected custom response: %#v", payload)
	}
	if payload.Palette.Primary == "#E63946" {
		t.Fatalf("expected custom primary color to be adapted")
	}
}

func TestQuickTestIsOrientative(t *testing.T) {
	body := []byte(`{"answers":{"reds_look_darker":false,"green_brown_confusion":false,"blue_yellow_confusion":true,"colors_look_gray":false}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/suggest", bytes.NewReader(body))
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.QuickTestResponse
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if payload.SuggestedType != "tritanopia" {
		t.Fatalf("unexpected suggestion: %#v", payload)
	}
	if payload.Disclaimer == "" {
		t.Fatal("missing medical disclaimer")
	}
}

func TestInvalidTypeContract(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/protanomaly", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}

	var payload map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	want := map[string]string{"error": "invalid_type", "message": "Tipo de daltonismo no soportado"}
	if !reflect.DeepEqual(payload, want) {
		t.Fatalf("unexpected error response: %#v", payload)
	}
}

func TestLegacyEndpointRemoved(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulate", nil)
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}
