package registrationlinksgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Create makes a new parent registration invite for one group,
// director-only. The returned Pin is the only place the raw PIN is ever
// exposed — staff reads it off screen to hand to the parent in person or
// over the phone; only its hash is stored.
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request registrationlinkspg.CreateRequest
		res     registrationlinkspg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "registrationlinksgt.Create", r.Body)
	defer logger.Exit(ctx, "registrationlinksgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
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

	inOrg, err := groupInOrganization(ctx, request.GroupId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, errGroupNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	res, err = registrationlinks.Create(ctx, registrationlinkspg.CreateArgument{
		OrganizationId:    organizationId,
		GroupId:           request.GroupId,
		ChildrenCount:     request.ChildrenCount,
		PublicDescription: request.PublicDescription,
		InternalNote:      request.InternalNote,
		CreatedBy:         accountId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
