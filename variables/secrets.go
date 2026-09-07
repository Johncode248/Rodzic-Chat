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

var STRIPESecret = mustEnv("STRIPE_SECRET_KEY")
var STRIPEWebhook = mustEnv("STRIPE_WEBHOOK_SECRET")

var JWTSecret = mustEnv("JWT_SECRET")

var SMTPHost = envOr("SMTP_HOST", "smtp-relay.brevo.com")
var SMTPPort = envOr("SMTP_PORT", "587")
var SMTPLogin = mustEnv("SMTP_LOGIN")
var SMTPPassword = mustEnv("SMTP_PASSWORD")
var SMTPFrom = envOr("SMTP_FROM", "noreply@dom2sky.pl")

// PlatformBankAccountNumber/Holder are shown to every director in the
// "Rozliczenie platformy" screen as the destination for their manual bank
// transfer -- same account for every organization, distinguished only by
// each organization's own PaymentReference. Not secret, but must be set to
// a real account before going live -- the fallback values are obvious
// placeholders on purpose.
var PlatformBankAccountNumber = envOr("PLATFORM_BANK_ACCOUNT_NUMBER", "UZUPEŁNIJ_NUMER_KONTA_W_ENV")
var PlatformBankAccountHolder = envOr("PLATFORM_BANK_ACCOUNT_HOLDER", "Flashy Code Stanisław Marcisz")
