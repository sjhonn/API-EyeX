package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/eyex-api/eyex/internal/models"
	"github.com/eyex-api/eyex/internal/theme"
)

type API struct {
	webDir string
}

func New(webDir string) *API {
	return &API{webDir: webDir}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/theme/types", a.types)
	mux.HandleFunc("GET /api/v1/theme/{type}", a.theme)

	if strings.TrimSpace(a.webDir) != "" {
		fs := http.FileServer(http.Dir(a.webDir))
		mux.Handle("GET /assets/", fs)
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
	result, ok := theme.Get(r.PathValue("type"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_type",
			Message: "Tipo de daltonismo no soportado",
		})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
