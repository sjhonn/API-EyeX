package handlers

import (
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
	want := []string{"normal", "protanopia", "deuteranopia", "tritanopia", "achromatopsia"}
	if !reflect.DeepEqual(payload.Types, want) {
		t.Fatalf("unexpected types: %#v", payload.Types)
	}
}

func TestDeuteranopiaContract(t *testing.T) {
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
	if payload.Palette.Background != "#1E1E1E" || payload.Palette.Primary != "#4A90D9" || payload.Palette.Success != "#4AD98C" {
		t.Fatalf("unexpected palette: %#v", payload.Palette)
	}
}

func TestAllSupportedThemes(t *testing.T) {
	for _, typeValue := range []string{"normal", "protanopia", "deuteranopia", "tritanopia", "achromatopsia"} {
		t.Run(typeValue, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/theme/"+typeValue, nil)
			res := httptest.NewRecorder()
			New("").Routes().ServeHTTP(res, req)
			if res.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
			}
		})
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
	want := map[string]string{
		"error":   "invalid_type",
		"message": "Tipo de daltonismo no soportado",
	}
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
