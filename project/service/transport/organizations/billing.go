package organizationsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/variables"
)

// Billing returns everything a director needs to pay (or check) the
// platform subscription: the fixed bank account details, this
// organization's own generated transfer-title reference, and the price
// computed from its current ChildrenLimit tier.
func Billing(w http.ResponseWriter, r *http.Request) {
	var (
		res organizationspg.BillingResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.Billing")
	defer logger.Exit(ctx, "organizationsgt.Billing", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	result, err := organizations.Select(ctx, organizationspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: organizationId}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(result.Organizations) == 0 {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	organization := result.Organizations[0]
	price, _ := organizationspg.MonthlyPriceGrosze(organization.ChildrenLimit)

	res = organizationspg.BillingResponse{
		ChildrenLimit:            organization.ChildrenLimit,
		MonthlyPriceGrosze:       price,
		PaymentReference:         organization.PaymentReference,
		SubscriptionFreeOverride: organization.SubscriptionFreeOverride,
		BankAccountNumber:        variables.PlatformBankAccountNumber,
		BankAccountHolder:        variables.PlatformBankAccountHolder,
	}

	json.NewEncoder(w).Encode(res)
}
