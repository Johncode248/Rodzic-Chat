package variables

import (
	"log"
	"os"
)

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("missing required environment variable: %s", name)
	}
	return value
}

func envOr(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

var JWTSecret = mustEnv("JWT_SECRET")

var SMTPHost = envOr("SMTP_HOST", "smtp-relay.brevo.com")
var SMTPPort = envOr("SMTP_PORT", "587")
var SMTPLogin = mustEnv("SMTP_LOGIN")
var SMTPPassword = mustEnv("SMTP_PASSWORD")
var SMTPFrom = envOr("SMTP_FROM", "noreply@dom2sky.pl")
