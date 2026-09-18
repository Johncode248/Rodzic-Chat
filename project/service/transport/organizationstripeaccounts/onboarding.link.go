package organizationstripeaccountsgt

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationstripeaccountspg "github.com/Johncode248/Rodzic-Chat/project/common/organizationstripeaccounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizationstripeaccounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// OnboardingLink handles POST /private/organizations/stripe/onboarding-link
// (director/co-director only, gated in module.go): creates the org's
// Stripe Connect Express account on first call, and returns a fresh
// Stripe-hosted onboarding (or, once already onboarded, "update details")
// link either way.
func OnboardingLink(w http.ResponseWriter, r *http.Request) {
	var (
		res organizationstripeaccountspg.OnboardingLinkResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "organizationstripeaccountsgt.OnboardingLink")
	defer logger.Exit(ctx, "organizationstripeaccountsgt.OnboardingLink", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	returnURL := fmt.Sprintf("%s/settings/organization?stripe=return", variables.FrontendURL)
	refreshURL := fmt.Sprintf("%s/settings/organization?stripe=refresh", variables.FrontendURL)

	res, err = organizationstripeaccounts.EnsureOnboardingLink(ctx, organizationId, returnURL, refreshURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
