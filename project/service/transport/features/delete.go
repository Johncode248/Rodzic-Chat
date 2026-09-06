package featuresgt

import (
	"encoding/json"
	"net/http"
	"strconv"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete removes a feature row for the caller's own organization, identified
// by ?feature= query param.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res featurespg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "featuresgt.Delete")
	defer logger.Exit(ctx, "featuresgt.Delete", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	featureOrdinal, err := strconv.Atoi(r.URL.Query().Get("feature"))
	if err != nil {
		http.Error(w, "Invalid or missing 'feature' query param", http.StatusBadRequest)
		return
	}

	res, err = features.Delete(ctx, featurespg.DeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "organization_id", Operator: "=", Value: organizationId},
						{Key: "feature", Operator: "=", Value: featureOrdinal},
					},
				},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
