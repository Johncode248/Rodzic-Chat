package organizationsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/organizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Update lets a director update only their own organization: the target id
// is forced to the caller's JWT organizationId, never a client-supplied one.
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request organizationspg.UpdateRequest
		res     organizationspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "organizationsgt.Update", r.Body)
	defer logger.Exit(ctx, "organizationsgt.Update", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = organizations.Update(ctx, organizationspg.UpdateArgument{
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
		Name:          request.Name,
		Address:       request.Address,
		Nip:           request.Nip,
		LogoUrl:       request.LogoUrl,
		ChildrenLimit: request.ChildrenLimit,
		PrimaryColor:  request.PrimaryColor,
	})
	if err != nil {
		if errors.Is(err, organizations.ErrChildrenLimitBelowCurrentCount) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
