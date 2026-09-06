package guardianclaimsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	guardianclaimspg "github.com/Johncode248/Rodzic-Chat/project/common/guardianclaims"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/guardianclaims"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Select lists guardian claims in the caller's organization, director-only.
// Defaults to only pending claims (?status=all to see resolved ones too) --
// this is the review queue for the "second parent matched by name but the
// child's birth date didn't agree" security check (see
// project/common/guardianclaims).
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res guardianclaimspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "guardianclaimsgt.Select")
	defer logger.Exit(ctx, "guardianclaimsgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{Groups: []databasepg.Group{{}}}

	if r.URL.Query().Get("status") != "all" {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "status", Operator: "=", Value: guardianclaimspg.GuardianClaimStatusPending.Ordinal(),
		})
	}

	query = query.AddOrganizationAuth(organizationId)

	res, err = guardianclaims.Select(ctx, guardianclaimspg.SelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
