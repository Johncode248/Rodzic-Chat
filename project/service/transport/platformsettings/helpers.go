package platformsettingsgt

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// requirePlatformAdminSecret is the whole "auth" story for these two
// endpoints: there is no platform-superadmin account type in this system
// (see the product decision in project/service/internal/authorization --
// the platform fee is a single global value the operator changes by hand,
// not a per-organization admin panel), so access is gated by a shared
// secret in the PLATFORM_ADMIN_SECRET env var instead. If that env var is
// unset, these endpoints refuse every request -- there's no sensible
// default secret to fall back to.
func requirePlatformAdminSecret(r *http.Request) bool {
	want := os.Getenv("PLATFORM_ADMIN_SECRET")
	if want == "" {
		return false
	}
	got := r.Header.Get("X-Platform-Admin-Secret")
	return subtle.ConstantTimeCompare([]byte(want), []byte(got)) == 1
}
