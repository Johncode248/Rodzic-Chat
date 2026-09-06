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

// Select returns the caller's own organization. Organizations have no
// organization_id column of their own, so instead of AddOrganizationAuth we
// scope by matching the record's id against the caller's JWT organizationId.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res organizationspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.Select")
	defer logger.Exit(ctx, "organizationsgt.Select", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusBadRequest)
		return
	}

	res, err = organizations.Select(ctx, organizationspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
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
