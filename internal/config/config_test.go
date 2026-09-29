package config

import (
	"strings"
	"testing"
)

// setRequired sets what Load refuses to start without, and clears every
// variable these tests read so the developer's own .env cannot leak in.
func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/fleet")
	t.Setenv("JWT_SECRET", "secret")
	for _, k := range []string{
		"SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASSWORD", "SMTP_FROM",
		"REPORTS_ENABLED", "REPORTS_TICK", "REPORTS_SEND_HOUR",
	} {
		t.Setenv(k, "")
	}
}

func TestReportsDefaults(t *testing.T) {
	setRequired(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Reports.Enabled {
		t.Error("the scheduler must be off unless REPORTS_ENABLED says otherwise")
	}
	if cfg.Reports.Tick != "*/15 * * * *" || cfg.Reports.SendHour != 6 {
		t.Errorf("reports = %+v", cfg.Reports)
	}
	if cfg.SMTP.Host != "" || cfg.SMTP.Port != 587 {
		t.Errorf("smtp = %+v", cfg.SMTP)
	}
}

func TestReportsFromEnv(t *testing.T) {
	setRequired(t)
	t.Setenv("REPORTS_ENABLED", "true")
	t.Setenv("REPORTS_TICK", "@hourly")
	t.Setenv("REPORTS_SEND_HOUR", "8")
	t.Setenv("SMTP_HOST", " mail.example.test ")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_FROM", "reportes@example.test")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Reports.Enabled || cfg.Reports.Tick != "@hourly" || cfg.Reports.SendHour != 8 {
		t.Errorf("reports = %+v", cfg.Reports)
	}
	if cfg.SMTP.Host != "mail.example.test" || cfg.SMTP.Port != 465 || cfg.SMTP.From != "reportes@example.test" {
		t.Errorf("smtp = %+v", cfg.SMTP)
	}
}

func TestReportsConfigIsValidated(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"a mail server with no sender", map[string]string{"SMTP_HOST": "mail.example.test"}, "SMTP_FROM"},
		{"port out of range", map[string]string{"SMTP_PORT": "70000"}, "SMTP_PORT"},
		{"port zero", map[string]string{"SMTP_PORT": "0"}, "SMTP_PORT"},
		{"hour 24", map[string]string{"REPORTS_SEND_HOUR": "24"}, "REPORTS_SEND_HOUR"},
		{"negative hour", map[string]string{"REPORTS_SEND_HOUR": "-1"}, "REPORTS_SEND_HOUR"},
		{"from with a display name", map[string]string{"SMTP_FROM": "Flota <x@y.mx>"}, "SMTP_FROM"},
		{"from with two addresses", map[string]string{"SMTP_FROM": "x@y.mx, z@y.mx"}, "SMTP_FROM"},
		{"from with no at sign", map[string]string{"SMTP_FROM": "no-at"}, "SMTP_FROM"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequired(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want it to name %s", err, tt.want)
			}
		})
	}
}

// A value that is not a number falls back to the default, as the other env
// helpers do, rather than stopping the process.
func TestNonNumericValuesFallBack(t *testing.T) {
	setRequired(t)
	t.Setenv("SMTP_PORT", "quinientos")
	t.Setenv("REPORTS_SEND_HOUR", "seis")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SMTP.Port != 587 || cfg.Reports.SendHour != 6 {
		t.Errorf("port = %d, hour = %d", cfg.SMTP.Port, cfg.Reports.SendHour)
	}
}
