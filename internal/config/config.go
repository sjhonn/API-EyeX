package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address            string
	AllowedOrigin      string
	Environment        string
	APIKey             string
	RateLimitPerMinute int
	RequestTimeout     time.Duration
}

func Load(path string) (Config, error) {
	if err := loadDotEnv(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	port, err := envInt("EYEX_PORT", 8080, 1, 65535)
	if err != nil {
		return Config{}, err
	}
	rateLimit, err := envInt("EYEX_RATE_LIMIT_PER_MINUTE", 60, 1, 1_000_000)
	if err != nil {
		return Config{}, err
	}
	timeoutMS, err := envInt("EYEX_REQUEST_TIMEOUT_MS", 5000, 100, 300_000)
	if err != nil {
		return Config{}, err
	}
	environment := strings.ToLower(envOr("EYEX_ENV", "development"))
	if environment != "development" && environment != "production" && environment != "test" {
		return Config{}, fmt.Errorf("EYEX_ENV must be development, production or test")
	}

	return Config{
		Address:            fmt.Sprintf(":%d", port),
		AllowedOrigin:      envOr("EYEX_ALLOWED_ORIGIN", "*"),
		Environment:        environment,
		APIKey:             strings.TrimSpace(os.Getenv("EYEX_API_KEY")),
		RateLimitPerMinute: rateLimit,
		RequestTimeout:     time.Duration(timeoutMS) * time.Millisecond,
	}, nil
}

func envInt(key string, fallback, minValue, maxValue int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minValue || value > maxValue {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", key, minValue, maxValue)
	}
	return value, nil
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
