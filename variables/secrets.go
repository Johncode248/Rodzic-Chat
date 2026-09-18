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

// mustEnvLen is mustEnv plus a required byte length, for values that would
// otherwise fail silently at first use rather than at boot -- e.g.
// FieldEncryptionKey below, which aes.NewCipher only rejects (with an
// error easy to miss deep in a request path) the first time something
// actually tries to encrypt/decrypt a PESEL, potentially long after a
// misconfigured deploy went live.
func mustEnvLen(name string, length int) string {
	value := mustEnv(name)
	if len(value) != length {
		log.Fatalf("environment variable %s must be exactly %d bytes, got %d", name, length, len(value))
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

// GoogleOAuthClientId/Secret back the web-based Google login redirect flow
// (project/service/transport/authorization/login.google.go). This is also
// the "audience" Android's native Google Sign-In must be configured to
// request an id_token for (its "server client ID") -- see
// project/service/internal/oauthverify/google.go -- so the backend can
// verify both the web redirect flow and Android's native flow against one
// known value. Previously hardcoded in source; moved to env because a
// Client Secret has no business being in version control.
var GoogleOAuthClientId = envOr("GOOGLE_OAUTH_CLIENT_ID", "847493899569-93j32dc40hev2dqtfc150sg1d0vdjupd.apps.googleusercontent.com")

// GoogleOAuthClientSecret has no safe placeholder -- unlike the client ID
// (public by design), a client secret must never ship a real value in
// source. Previously had a real secret hardcoded as the fallback; that
// secret must be rotated in the Google Cloud console.
var GoogleOAuthClientSecret = mustEnv("GOOGLE_OAUTH_CLIENT_SECRET")

// AppleAudience is the iOS app's bundle identifier -- what native Sign in
// with Apple puts in an identity token's "aud" claim (see
// project/service/internal/oauthverify/apple.go). Must be set to the real
// bundle id before Apple Sign-In can work; the fallback is an obvious
// placeholder.
var AppleAudience = envOr("APPLE_AUDIENCE", "UZUPEŁNIJ_BUNDLE_ID_W_ENV")

// AllowTestEndpoints gates scaffolding that must never be reachable in
// production (currently just POST /private/parent-subscription/test-set-plan
// -- see its own doc comment). Defaults to OFF (safe-by-default): a
// deployment must explicitly opt in with ALLOW_TEST_ENDPOINTS=true (local
// dev / staging only) rather than the endpoint being open unless someone
// remembers to lock it down.
var AllowTestEndpoints = envOr("ALLOW_TEST_ENDPOINTS", "false") == "true"

// CORSAllowedOrigins is a comma-separated list of extra origins (e.g. a
// local dev frontend on http://localhost:3000) allowed to make
// credentialed cross-origin requests, on top of DOMEIN_F which is always
// allowed -- see CorsMiddleware. Empty by default.
var CORSAllowedOrigins = envOr("CORS_ALLOWED_ORIGINS", "")

// AnthropicApiKey backs the parent-facing AI assistant (see
// project/service/internal/aichat). Deliberately NOT a mustEnv() -- unlike
// Stripe/JWT/SMTP, the rest of the app must keep working (parents can still
// see their plan/limits, everything else keeps functioning) while this is
// unset; project/service/internal/aichat.Send checks for an empty key
// itself and returns a clear "not configured yet" error rather than ever
// calling Anthropic with one.
var AnthropicApiKey = envOr("ANTHROPIC_API_KEY", "")

// AnthropicModel: Haiku 4.5 -- chosen for cost/speed given this is a
// high-volume, simple-recommendation use case (see product discussion),
// not Sonnet/Opus. Override via env without a code change if that
// calculus changes later.
var AnthropicModel = envOr("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001")

// GooglePlacesApiKey backs the PRIMARY place search (project/service/internal/places.SearchGoogle
// via SearchWithFallback) -- Google's broader coverage of small local businesses (doctors,
// specialists) that OpenStreetMap often lacks. Every parent tries Google first now (not gated by
// plan -- MVP only has PlanFree), bounded by a global monthly request budget plus a per-parent
// daily cap (see placespg.MonthlyGooglePlacesRequestBudget and
// parentsubscriptionspg.DailyGooglePlacesSearchLimit); either budget running out transparently
// falls back to the always-free OpenStreetMap search instead of failing. Deliberately NOT a
// mustEnv() -- like AnthropicApiKey, everything else keeps working (OSM search stays the
// fallback) while this is unset; SearchGoogle checks for an empty key itself and returns a clear
// "not configured yet" error, which SearchWithFallback treats as just another reason to fall
// back to OSM.
var GooglePlacesApiKey = envOr("GOOGLE_PLACES_API_KEY", "")

// FieldEncryptionKey encrypts individual sensitive DB columns at rest
// (currently just Child.EncryptedPesel -- see library/crypto) with
// AES-256-GCM, which requires exactly 32 bytes. No fallback -- unlike
// AnthropicApiKey/GooglePlacesApiKey, missing this key must fail the whole
// process at startup rather than silently encrypting PESELs with a
// hardcoded, publicly-known-from-source key. Production must set
// FIELD_ENCRYPTION_KEY to a real random 32-byte value (e.g. `openssl rand
// -base64 32 | cut -c1-32`) and never let it change once real PESELs are
// stored under it, or every existing one becomes permanently
// undecryptable.
var FieldEncryptionKey = mustEnvLen("FIELD_ENCRYPTION_KEY", 32)

// R2AccessKey/R2SecretKey/R2Endpoint back Cloudflare R2 object storage (see
// api_bucket.Init). Previously hardcoded in source; a real key was leaked
// that way and must be rotated in the Cloudflare dashboard.
var R2AccessKey = mustEnv("R2_ACCESS_KEY")
var R2SecretKey = mustEnv("R2_SECRET_KEY")
var R2Endpoint = mustEnv("R2_ENDPOINT")

// DatabaseURL was previously hardcoded directly in database/gormdb/module.go
// as a personal dev machine's local, passwordless Postgres -- unlike every
// other piece of config in this file, that meant no deployment could point
// at a different database (staging, prod, a rotated password) without
// recompiling the binary. envOr, not mustEnv: the fallback keeps today's
// local dev experience working unchanged for anyone who doesn't set
// DATABASE_URL.
var DatabaseURL = mustEnv("DATABASE_URL")

// "host=localhost port=5432 user=staszek password= dbname=staszek sslmode=disable connect_timeout=15"

var BackendURL = envOr("BACKEND_URL", "http://localhost:8080")

var FrontendURL = envOr("FRONTEND_URL", "https://platforma.rodzicchat.pl")
