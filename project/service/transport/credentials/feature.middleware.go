package credentials

import (
	"net/http"

	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
)

// RequireFeature returns a middleware that only allows requests through when
// the given feature is enabled for the caller's organization (per the JWT's
// organizationId). Compose it with RequireAccountType where both apply, e.g.
// staffOnly(RequireFeature(featurespg.FeatureChat)(http.HandlerFunc(...))).
func RequireFeature(feature featurespg.Feature) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			organizationId, err := GetOrganizationIdFromContext(r.Context())
			if err != nil || organizationId == "" {
				http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
				return
			}

			enabled, err := features.IsEnabled(r.Context(), organizationId, feature)
			if err != nil {
				http.Error(w, "Error checking feature availability", http.StatusInternalServerError)
				return
			}

			if !enabled {
				http.Error(w, "feature not enabled for this organization", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
