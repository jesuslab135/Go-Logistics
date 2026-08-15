package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Env         string
	HTTPAddr    string
	DatabaseURL string
	CORSOrigins []string
	Swagger     bool
	JWT         JWTConfig
}

type JWTConfig struct {
	Secret     string
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Load reads configuration from the environment. DATABASE_URL and JWT_SECRET are
// required; everything else has a development-friendly default.
func Load() (Config, error) {
	cfg := Config{
		Env:         env("APP_ENV", "development"),
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CORSOrigins: splitCSV(env("CORS_ORIGINS", "*")),
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			Issuer:     env("JWT_ISSUER", "fleet"),
			AccessTTL:  envDuration("JWT_ACCESS_TTL", time.Hour),
			RefreshTTL: envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
	}

	// Serving the docs publishes the entire API surface, so production defaults
	// to off — but that is a judgement call, not a safety property, and it is
	// separate from APP_ENV so turning the docs on does not also drop the app
	// out of Gin's release mode.
	cfg.Swagger = envBool("SWAGGER_ENABLED", !cfg.IsProduction())

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: DATABASE_URL is required")
	}
	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("config: JWT_SECRET is required")
	}
	return cfg, nil
}

func (c Config) IsProduction() bool { return c.Env == "production" }

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// envBool accepts the values people actually type. Anything unrecognised falls
// back rather than silently reading as false.
func envBool(key string, fallback bool) bool {
	switch strings.ToLower(env(key, "")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
