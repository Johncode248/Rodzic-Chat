package organizationsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete lets a director delete only their own organization (target id is
// forced to the caller's JWT organizationId).
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res organizationspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.Delete")
	defer logger.Exit(ctx, "organizationsgt.Delete", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = organizations.Delete(ctx, organizationspg.DeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{
							Key:      "id",
							Operator: "=",
							Value:    organizationId,
						},
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
