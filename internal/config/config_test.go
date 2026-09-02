package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("EYEX_PORT", "")
	t.Setenv("EYEX_ALLOWED_ORIGIN", "")
	t.Setenv("EYEX_ENV", "")
	t.Setenv("EYEX_API_KEY", "")
	t.Setenv("EYEX_RATE_LIMIT_PER_MINUTE", "")
	t.Setenv("EYEX_REQUEST_TIMEOUT_MS", "")
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != ":8080" || cfg.RateLimitPerMinute != 60 || cfg.RequestTimeout != 5*time.Second || cfg.Environment != "development" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadSingleDotEnv(t *testing.T) {
	keys := []string{"EYEX_PORT", "EYEX_ALLOWED_ORIGIN", "EYEX_ENV", "EYEX_API_KEY", "EYEX_RATE_LIMIT_PER_MINUTE", "EYEX_REQUEST_TIMEOUT_MS"}
	for _, key := range keys {
		old, ok := os.LookupEnv(key)
		_ = os.Unsetenv(key)
		if ok {
			defer os.Setenv(key, old)
		}
	}
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("EYEX_PORT=9090\nEYEX_ENV=test\nEYEX_RATE_LIMIT_PER_MINUTE=12\nEYEX_REQUEST_TIMEOUT_MS=900\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != ":9090" || cfg.Environment != "test" || cfg.RateLimitPerMinute != 12 || cfg.RequestTimeout != 900*time.Millisecond {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
