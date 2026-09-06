package registrationlinksgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete revokes a registration link identified by ?id=, director-only,
// org-scoped. This is a hard delete — a revoked link's token stops working
// immediately since it's simply gone.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res registrationlinkspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "registrationlinksgt.Delete")
	defer logger.Exit(ctx, "registrationlinksgt.Delete", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}},
		},
	}
	query = query.AddOrganizationAuth(organizationId)

	res, err = registrationlinks.Delete(ctx, registrationlinkspg.DeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
