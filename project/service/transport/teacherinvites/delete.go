package teacherinvitesgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete revokes a teacher invite identified by ?id=, director-only,
// org-scoped. Hard delete — a revoked invite's token stops working
// immediately since it's simply gone.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res teacherinvitespg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.Delete")
	defer logger.Exit(ctx, "teacherinvitesgt.Delete", &res, &err)

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

	res, err = teacherinvites.Delete(ctx, teacherinvitespg.DeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
