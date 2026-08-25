package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address       string
	AllowedOrigin string
}

func Load(path string) (Config, error) {
	if err := loadDotEnv(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	port := 8080
	if raw := strings.TrimSpace(os.Getenv("EYEX_PORT")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 65535 {
			return Config{}, fmt.Errorf("EYEX_PORT must be an integer between 1 and 65535")
		}
		port = value
	}

	return Config{
		Address:       fmt.Sprintf(":%d", port),
		AllowedOrigin: envOr("EYEX_ALLOWED_ORIGIN", "*"),
	}, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
