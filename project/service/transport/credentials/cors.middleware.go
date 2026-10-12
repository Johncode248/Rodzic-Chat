package credentials

import (
	"net/http"
	"os"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/variables"
)

// allowedOrigins is the set of origins permitted to make credentialed
// cross-origin requests -- DOMEIN_F (the real web frontend) plus whatever
// variables.CORSAllowedOrigins adds (e.g. a local dev frontend).
//
// "Access-Control-Allow-Origin: *" together with
// "Access-Control-Allow-Credentials: true" is invalid per the Fetch spec
// (a wildcard origin is not allowed alongside credentials) -- browsers
// reject the response outright, which previously broke every credentialed
// request from the web frontend. Reflecting a specific, allowlisted Origin
// back (with Vary: Origin so caches don't mix responses for different
// origins) is the standard fix.
var allowedOrigins = buildAllowedOrigins()

func buildAllowedOrigins() map[string]bool {
	origins := map[string]bool{}
	if variables.FrontendURL != "" {
		origins[variables.FrontendURL] = true
	}
	for _, o := range strings.Split(variables.CORSAllowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins[o] = true
		}
	}
	return origins
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Local admin console (only mounted with ADMIN_MODE=1, and loopback-only inside the
		// handler): allow any origin -- including "null" for an HTML file opened from disk.
		adminLocal := os.Getenv("ADMIN_MODE") == "1" && strings.HasPrefix(r.URL.Path, "/admin/api/")
		if origin != "" && (allowedOrigins[origin] || adminLocal) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		w.Header().Set("Content-Type", "application/json")

		// Obsługa preflight (OPTIONS)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
