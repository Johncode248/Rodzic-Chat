package teacherinvitesgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errOnlyDirectorCanInviteCoDirector = errors.New("only a director (not a co-director) can invite another co-director")

// Create makes a new teacher invite link, director-only. The returned
// Token is the only place the raw token is ever exposed — the caller (the
// staff web platform) builds the shareable URL itself.
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request teacherinvitespg.CreateRequest
		res     teacherinvitespg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.Create", r.Body)
	defer logger.Exit(ctx, "teacherinvitesgt.Create", &res, &err)

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

	if request.AccountType == "co_director" {
		callerAccountType, err := credentials.GetAccountTypeFromContext(r.Context())
		if err != nil || callerAccountType != accountspg.AccountTypeDirector.Ordinal() {
			http.Error(w, errOnlyDirectorCanInviteCoDirector.Error(), http.StatusForbidden)
			return
		}
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

	res, err = teacherinvites.Create(ctx, teacherinvitespg.CreateArgument{
		OrganizationId:    organizationId,
		GroupId:           request.GroupId,
		AccountType:       request.AccountType,
		Name:              request.Name,
		Email:             request.Email,
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
