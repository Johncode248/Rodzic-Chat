package organizationsgt

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"

	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
)

// requirePlatformAdminSecret gates SetSubscriptionOverride the same way
// project/service/transport/platformsettings gates its endpoints: there is
// no platform-superadmin account type in this system, so access is a
// shared secret in the PLATFORM_ADMIN_SECRET env var instead of a JWT role.
func requirePlatformAdminSecret(r *http.Request) bool {
	want := os.Getenv("PLATFORM_ADMIN_SECRET")
	if want == "" {
		return false
	}
	got := r.Header.Get("X-Platform-Admin-Secret")
	return subtle.ConstantTimeCompare([]byte(want), []byte(got)) == 1
}

// SetSubscriptionOverride lets the platform operator exempt one
// organization from the subscription fee entirely (or revoke that
// exemption) -- never reachable by a director themselves. Mounted under
// /public since it authenticates via shared secret rather than a JWT,
// matching project/service/transport/platformsettings.
func SetSubscriptionOverride(w http.ResponseWriter, r *http.Request) {
	var (
		request organizationspg.SubscriptionOverrideRequest
		res     organizationspg.SubscriptionOverrideResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.SetSubscriptionOverride", r.Body)
	defer logger.Exit(ctx, "organizationsgt.SetSubscriptionOverride", &res, &err)

	if !requirePlatformAdminSecret(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	if request.OrganizationId == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
		return
	}

	organization, err := organizations.SetSubscriptionFreeOverride(ctx, request.OrganizationId, request.FreeOverride)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = organizationspg.SubscriptionOverrideResponse{Organization: organization}
	json.NewEncoder(w).Encode(res)
}
