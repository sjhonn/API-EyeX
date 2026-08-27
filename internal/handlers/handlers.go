package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/eyex-api/eyex/internal/models"
	"github.com/eyex-api/eyex/internal/theme"
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
	mux.HandleFunc("GET /api/v1/theme/{type}", a.theme)

	if strings.TrimSpace(a.webDir) != "" {
		fs := http.FileServer(http.Dir(a.webDir))
		mux.Handle("GET /assets/", fs)
		mux.HandleFunc("GET /eyex.css", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(a.webDir, "eyex.css"))
		})
		mux.HandleFunc("GET /eyex.js", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(a.webDir, "eyex.js"))
		})
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(a.webDir, "index.html"))
		})
	}
	return mux
}

func (a *API) types(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, models.TypesResponse{Types: theme.Types()})
}

func (a *API) theme(w http.ResponseWriter, r *http.Request) {
	options, err := queryOptions(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_parameter", Message: err.Error()})
		return
	}
	result, ok, err := theme.Get(r.PathValue("type"), options)
	if !ok {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_type", Message: "Tipo de daltonismo no soportado"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_parameter", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) customTheme(w http.ResponseWriter, r *http.Request) {
	var req models.CustomThemeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "JSON de entrada inválido"})
		return
	}
	result, err := theme.Custom(req)
	if err != nil {
		if err.Error() == "invalid_type" {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_type", Message: "Tipo de daltonismo no soportado"})
			return
		}
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_palette", Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *API) quickTest(w http.ResponseWriter, r *http.Request) {
	var req models.QuickTestRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "JSON de entrada inválido"})
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
