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

func TestSimulateMachadoEndpoint(t *testing.T) {
	body := []byte(`{"hex":"#ff0000","type":"protanopia","severity":0.65}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulate", bytes.NewReader(body))
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.SimulateResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Original != "#FF0000" || payload.Simulated != "#A05A00" || payload.Model != "machado-2009" {
		t.Fatalf("unexpected simulation: %#v", payload)
	}
	if payload.Type != "protanopia" || payload.Severity != 0.65 {
		t.Fatalf("unexpected metadata: %#v", payload)
	}
}

func TestSimulateDefaultsToFullSeverity(t *testing.T) {
	body := []byte(`{"hex":"#FF0000","type":"protanopia"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulate", bytes.NewReader(body))
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.SimulateResponse
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if payload.Severity != 1 || payload.Simulated != "#6D5F00" {
		t.Fatalf("unexpected default simulation: %#v", payload)
	}
}

func TestSimulateBatch(t *testing.T) {
	body := []byte(`{"colors":["#FF0000","#00ff00","#0000FF"],"type":"deuteranopia","severity":0.5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulate/batch", bytes.NewReader(body))
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.SimulateBatchResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Results) != 3 || payload.Results[1].Original != "#00FF00" {
		t.Fatalf("unexpected batch response: %#v", payload)
	}
}

func TestSimulateRejectsInvalidSeverity(t *testing.T) {
	body := []byte(`{"hex":"#336699","type":"tritanopia","severity":1.01}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulate", bytes.NewReader(body))
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.ErrorResponse
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if payload.Error != "invalid_parameter" {
		t.Fatalf("unexpected error: %#v", payload)
	}
}

func TestSimulateRejectsInvalidColor(t *testing.T) {
	body := []byte(`{"hex":"336699","type":"tritanopia","severity":0.4}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulate", bytes.NewReader(body))
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", res.Code, res.Body.String())
	}
	var payload models.ErrorResponse
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if payload.Error != "invalid_color" {
		t.Fatalf("unexpected error: %#v", payload)
	}
}

func TestThemeETagAndNotModified(t *testing.T) {
	h := New("").Routes()
	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/theme/deuteranopia", nil)
	first := httptest.NewRecorder()
	h.ServeHTTP(first, firstReq)
	etag := first.Header().Get("ETag")
	if first.Code != http.StatusOK || etag == "" || first.Header().Get("Cache-Control") == "" {
		t.Fatalf("missing cache headers: status=%d etag=%q cache=%q", first.Code, etag, first.Header().Get("Cache-Control"))
	}
	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/theme/deuteranopia", nil)
	secondReq.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	h.ServeHTTP(second, secondReq)
	if second.Code != http.StatusNotModified || second.Body.Len() != 0 {
		t.Fatalf("expected 304 without body, got %d %q", second.Code, second.Body.String())
	}
}

func TestRoutingErrorsAreJSON(t *testing.T) {
	h := New("").Routes()
	for _, tc := range []struct {
		method string
		path   string
		code   int
		error  string
	}{
		{http.MethodGet, "/api/v1/no-existe", http.StatusNotFound, "not_found"},
		{http.MethodDelete, "/api/v1/theme/types", http.StatusMethodNotAllowed, "method_not_allowed"},
	} {
		res := httptest.NewRecorder()
		h.ServeHTTP(res, httptest.NewRequest(tc.method, tc.path, nil))
		var payload models.ErrorResponse
		if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
			t.Fatalf("expected JSON for %s %s: %v / %s", tc.method, tc.path, err, res.Body.String())
		}
		if res.Code != tc.code || payload.Error != tc.error {
			t.Fatalf("unexpected routing error: status=%d payload=%#v", res.Code, payload)
		}
	}
}

func TestEnglishErrorMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/no-existe", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	res := httptest.NewRecorder()
	New("").Routes().ServeHTTP(res, req)
	var payload models.ErrorResponse
	_ = json.Unmarshal(res.Body.Bytes(), &payload)
	if payload.Message != "Unsupported color vision type" {
		t.Fatalf("unexpected translation: %#v", payload)
	}
}
