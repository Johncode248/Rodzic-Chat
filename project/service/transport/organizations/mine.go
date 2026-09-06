package organizationsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accountorganizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Mine lists every organization (facility) the caller belongs to -- for a
// parent with children at more than one placówka, that's every one of
// them, not just the JWT's current active org. Falls back to the caller's
// single active org when they have no AccountOrganization rows yet
// (pre-backfill), matching resolveOrganizations' identical fallback.
func Mine(w http.ResponseWriter, r *http.Request) {
	var (
		res organizationspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.Mine")
	defer logger.Exit(ctx, "organizationsgt.Mine", &res, &err)

	userId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Auth went wrong", http.StatusForbidden)
		return
	}

	memberships, err := accountorganizations.Select(ctx, accountorganizationspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: userId}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	organizationIds := make([]string, 0, len(memberships.AccountOrganizations))
	for _, membership := range memberships.AccountOrganizations {
		organizationIds = append(organizationIds, membership.OrganizationId)
	}

	if len(organizationIds) == 0 {
		organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
		if err != nil || organizationId == "" {
			http.Error(w, "Error getting organizationId from context", http.StatusBadRequest)
			return
		}
		organizationIds = append(organizationIds, organizationId)
	}

	res, err = organizations.Select(ctx, organizationspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: organizationIds}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
