package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/eyex-api/eyex/internal/models"
	"github.com/eyex-api/eyex/internal/theme"
	"github.com/eyex-api/eyex/pkg/colorblind"
)

const maxJSONBody = 32 << 10

type API struct {
	webDir string
}

func New(webDir string) *API {
	return &API{webDir: webDir}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/theme/types", a.types)
	mux.HandleFunc("POST /api/v1/theme/custom", a.customTheme)
	mux.HandleFunc("POST /api/v1/test/suggest", a.quickTest)
	mux.HandleFunc("POST /api/v1/simulate", a.simulate)
	mux.HandleFunc("POST /api/v1/simulate/batch", a.simulateBatch)
	mux.HandleFunc("GET /api/v1/theme/{type}", a.theme)
	mux.HandleFunc("GET /openapi.yaml", a.openAPI)

	if strings.TrimSpace(a.webDir) != "" {
		fs := http.FileServer(http.Dir(a.webDir))
		mux.Handle("GET /assets/", fs)
		mux.HandleFunc("GET /eyex.css", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(a.webDir, "eyex.css"))
		})
		mux.HandleFunc("GET /eyex.js", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(a.webDir, "eyex.js"))
		})
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(a.webDir, "index.html"))
		})
	}
	return normalizeRoutingErrors(mux)
}

func (a *API) openAPI(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat("openapi.yaml"); err != nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Recurso no encontrado")
		return
	}
	http.ServeFile(w, r, "openapi.yaml")
}

func normalizeRoutingErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)
		if recorder.Code == http.StatusNotFound {
			writeError(w, r, http.StatusNotFound, "not_found", "Recurso no encontrado")
			return
		}
		if recorder.Code == http.StatusMethodNotAllowed {
			if allow := recorder.Header().Get("Allow"); allow != "" {
				w.Header().Set("Allow", allow)
			}
			writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "Método no permitido")
			return
		}
		for key, values := range recorder.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(recorder.Code)
		_, _ = io.Copy(w, recorder.Body)
	})
}

func (a *API) types(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, models.TypesResponse{Types: theme.Types()})
}

func (a *API) theme(w http.ResponseWriter, r *http.Request) {
	options, err := queryOptions(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_parameter", err.Error())
		return
	}
	result, ok, err := theme.Get(r.PathValue("type"), options)
	if !ok {
		writeError(w, r, http.StatusBadRequest, "invalid_type", "Tipo de daltonismo no soportado")
		return
	}
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_parameter", err.Error())
		return
	}
	writeCacheableJSON(w, r, result)
}

func (a *API) customTheme(w http.ResponseWriter, r *http.Request) {
	var req models.CustomThemeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "JSON de entrada inválido")
		return
	}
	result, err := theme.Custom(req)
	if err != nil {
		if err.Error() == "invalid_type" {
			writeError(w, r, http.StatusBadRequest, "invalid_type", "Tipo de daltonismo no soportado")
			return
		}
		writeError(w, r, http.StatusBadRequest, "invalid_palette", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) simulate(w http.ResponseWriter, r *http.Request) {
	var req models.SimulateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "JSON de entrada inválido")
		return
	}
	deficiency, ok := simulationDeficiency(req.Type)
	if !ok {
		writeError(w, r, http.StatusBadRequest, "invalid_type", "Tipo de daltonismo no soportado")
		return
	}
	severity, err := simulationSeverity(req.Severity)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_parameter", err.Error())
		return
	}
	original, err := colorblind.NormalizeHex(req.Hex)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_color", "hex debe usar formato #RRGGBB")
		return
	}
	simulated, err := colorblind.SimulateHexSeverity(original, deficiency, severity)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_parameter", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.SimulateResponse{
		Original: original, Simulated: simulated, Type: req.Type, Severity: severity, Model: colorblind.MachadoModel,
	})
}

func (a *API) simulateBatch(w http.ResponseWriter, r *http.Request) {
	var req models.SimulateBatchRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "JSON de entrada inválido")
		return
	}
	if len(req.Colors) == 0 || len(req.Colors) > 256 {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "colors debe contener entre 1 y 256 colores")
		return
	}
	deficiency, ok := simulationDeficiency(req.Type)
	if !ok {
		writeError(w, r, http.StatusBadRequest, "invalid_type", "Tipo de daltonismo no soportado")
		return
	}
	severity, err := simulationSeverity(req.Severity)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_parameter", err.Error())
		return
	}
	results := make([]models.SimulatedColor, 0, len(req.Colors))
	for _, value := range req.Colors {
		original, err := colorblind.NormalizeHex(value)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid_color", "cada color debe usar formato #RRGGBB")
			return
		}
		simulated, err := colorblind.SimulateHexSeverity(original, deficiency, severity)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid_parameter", err.Error())
			return
		}
		results = append(results, models.SimulatedColor{Original: original, Simulated: simulated})
	}
	writeJSON(w, http.StatusOK, models.SimulateBatchResponse{
		Type: req.Type, Severity: severity, Model: colorblind.MachadoModel, Results: results,
	})
}

func simulationDeficiency(value string) (colorblind.Deficiency, bool) {
	switch value {
	case "protanopia":
		return colorblind.Protanopia, true
	case "deuteranopia":
		return colorblind.Deuteranopia, true
	case "tritanopia":
		return colorblind.Tritanopia, true
	default:
		return "", false
	}
}

func simulationSeverity(value *float64) (float64, error) {
	if value == nil {
		return 1, nil
	}
	if *value < 0 || *value > 1 {
		return 0, errors.New("severity debe estar entre 0 y 1")
	}
	return *value, nil
}

func (a *API) quickTest(w http.ResponseWriter, r *http.Request) {
	var req models.QuickTestRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "JSON de entrada inválido")
		return
	}
	writeJSON(w, http.StatusOK, models.QuickTestResponse{
		SuggestedType: theme.SuggestType(req.Answers),
		Disclaimer:    "Resultado orientativo. No es un diagnóstico médico.",
	})
}

func queryOptions(r *http.Request) (theme.Options, error) {
	q := r.URL.Query()
	severity := q.Get("severity")
	mode := q.Get("mode")
	highContrastRaw := q.Get("high_contrast")
	options := theme.Options{
		Severity: severity,
		Mode:     mode,
		Explicit: severity != "" || mode != "" || highContrastRaw != "",
	}
	if highContrastRaw != "" {
		switch highContrastRaw {
		case "true", "1":
			options.HighContrast = true
		case "false", "0":
			options.HighContrast = false
		default:
			return theme.Options{}, errors.New("high_contrast debe ser true o false")
		}
	}
	return options, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("se esperaba un único objeto JSON")
	}
	return nil
}

func writeCacheableJSON(w http.ResponseWriter, r *http.Request, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "internal_server_error", "Error interno del servidor")
		return
	}
	sum := sha256.Sum256(body)
	etag := fmt.Sprintf("\"%x\"", sum[:12])
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=300")
	if strings.TrimSpace(r.Header.Get("If-None-Match")) == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, spanish string) {
	message := localizedMessage(r.Header.Get("Accept-Language"), spanish)
	writeJSON(w, status, models.ErrorResponse{Error: code, Message: message})
}

func localizedMessage(acceptLanguage, spanish string) string {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(acceptLanguage)), "en") {
		return spanish
	}
	translations := map[string]string{
		"Tipo de daltonismo no soportado":            "Unsupported color vision type",
		"JSON de entrada inválido":                   "Invalid JSON request",
		"hex debe usar formato #RRGGBB":              "hex must use #RRGGBB format",
		"cada color debe usar formato #RRGGBB":       "each color must use #RRGGBB format",
		"severity debe estar entre 0 y 1":            "severity must be between 0 and 1",
		"colors debe contener entre 1 y 256 colores": "colors must contain between 1 and 256 colors",
		"high_contrast debe ser true o false":        "high_contrast must be true or false",
		"severity debe ser mild, moderate o severe":  "severity must be mild, moderate or severe",
		"mode debe ser dark o light":                 "mode must be dark or light",
		"Método no permitido":                        "Method not allowed",
		"Recurso no encontrado":                      "Resource not found",
		"Error interno del servidor":                 "Internal server error",
	}
	if translated, ok := translations[spanish]; ok {
		return translated
	}
	return spanish
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
