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

// Select lists every OrganizationFeature row for the caller's organization.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res featurespg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "featuresgt.Select")
	defer logger.Exit(ctx, "featuresgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{{}},
	}
	query = query.AddOrganizationAuth(organizationId)

	res, err = features.Select(ctx, featurespg.SelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
