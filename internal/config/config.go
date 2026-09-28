package config

import (
	"fmt"
	"os"
	"strconv"
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
	SMTP        SMTPConfig
	Reports     ReportsConfig
}

type JWTConfig struct {
	Secret     string
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// SMTPConfig is the mail server the scheduled reports are sent through. An
// empty Host means none is configured: reports are still built, and their runs
// recorded as not sent.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// ReportsConfig drives the in-process scheduler.
type ReportsConfig struct {
	// Enabled starts the scheduler. It defaults to off so that a development
	// or QA process pointed at real data does not email anyone by accident.
	Enabled bool
	// Tick is the cron expression of the single job that looks for due reports.
	Tick string
	// SendHour is the local hour, in each company's own timezone, at which a
	// finished period becomes due.
	SendHour int
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
		SMTP: SMTPConfig{
			Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
			Port:     envInt("SMTP_PORT", 587),
			User:     os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		},
		Reports: ReportsConfig{
			Enabled:  envBool("REPORTS_ENABLED", false),
			Tick:     env("REPORTS_TICK", "*/15 * * * *"),
			SendHour: envInt("REPORTS_SEND_HOUR", 6),
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
	if cfg.SMTP.Host != "" && cfg.SMTP.From == "" {
		return Config{}, fmt.Errorf("config: SMTP_FROM is required when SMTP_HOST is set")
	}
	if cfg.SMTP.Port < 1 || cfg.SMTP.Port > 65535 {
		return Config{}, fmt.Errorf("config: SMTP_PORT must be between 1 and 65535")
	}
	if cfg.Reports.SendHour < 0 || cfg.Reports.SendHour > 23 {
		return Config{}, fmt.Errorf("config: REPORTS_SEND_HOUR must be between 0 and 23")
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

// envInt falls back on anything that is not a whole number, as envBool and
// envDuration do for their types.
func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
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
