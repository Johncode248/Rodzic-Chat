package featuresgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Update flips a feature's Enabled flag for the caller's own organization.
// The target row is looked up by (organization_id, feature) using the
// caller's own JWT organizationId -- a director can never touch another
// organization's feature flags.
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request featurespg.UpdateRequest
		res     featurespg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "featuresgt.Update", r.Body)
	defer logger.Exit(ctx, "featuresgt.Update", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = features.Update(ctx, featurespg.UpdateArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "organization_id", Operator: "=", Value: organizationId},
						{Key: "feature", Operator: "=", Value: int(request.Feature)},
					},
				},
			},
		},
		Enabled: request.Enabled,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
