package organizationstripeaccountsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationstripeaccountspg "github.com/Johncode248/Rodzic-Chat/project/common/organizationstripeaccounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizationstripeaccounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Status handles GET /private/organizations/stripe/status (any
// authenticated org member -- staff need it to gate the "create payment"
// UI, a parent's app could show it too for transparency).
func Status(w http.ResponseWriter, r *http.Request) {
	var (
		res organizationstripeaccountspg.StatusResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "organizationstripeaccountsgt.Status")
	defer logger.Exit(ctx, "organizationstripeaccountsgt.Status", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	account, err := organizationstripeaccounts.Get(ctx, organizationId)
	if err != nil {
		if errors.Is(err, organizationstripeaccounts.ErrNoAccountYet) {
			json.NewEncoder(w).Encode(organizationstripeaccountspg.StatusResponse{Connected: false})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = organizationstripeaccountspg.StatusResponse{
		Connected:        true,
		ChargesEnabled:   account.ChargesEnabled,
		PayoutsEnabled:   account.PayoutsEnabled,
		DetailsSubmitted: account.DetailsSubmitted,
	}

	json.NewEncoder(w).Encode(res)
}
