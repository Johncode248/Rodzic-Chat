package guardianclaimsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/guardianclaims"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

type ResolveRequest struct {
	Id      string `json:"id,omitempty"`
	Approve bool   `json:"approve,omitempty"`
}

type ResolveResponse struct{}

// Resolve lets a director approve ("tak, to są rodzice" -- creates the
// ParentChild link with pickup rights) or reject ("nie, któraś z tych
// osób to nie rodzic" -- no link is ever created) a pending guardian
// claim, director-only, org-scoped.
func Resolve(w http.ResponseWriter, r *http.Request) {
	var (
		request ResolveRequest
		res     ResolveResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "guardianclaimsgt.Resolve", &request)
	defer logger.Exit(ctx, "guardianclaimsgt.Resolve", &res, &err)

	if request.Id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err = guardianclaims.Resolve(ctx, organizationId, request.Id, accountId, request.Approve); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}
