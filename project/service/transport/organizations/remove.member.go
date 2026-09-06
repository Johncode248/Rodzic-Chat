package organizationsgt

import (
	"context"
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountorganizationspg "github.com/Johncode248/Rodzic-Chat/project/common/accountorganizations"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accountorganizations"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

type RemoveMemberRequest struct {
	AccountId string `json:"account_id,omitempty"`
}

type RemoveMemberResponse struct{}

// RemoveMember lets a director remove a parent's access to the director's
// own organization -- it never touches that parent's membership in any
// other organization (see accountorganizations.RemoveMember). If this was
// the parent's LAST organization, it reads naturally as "delete this
// parent" from the director's perspective, so it cascades into a full
// account soft-delete (accounts.Delete already revokes sessions and
// hard-deletes the Credential).
func RemoveMember(w http.ResponseWriter, r *http.Request) {
	var (
		request RemoveMemberRequest
		res     RemoveMemberResponse
		err     error
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	ctx := logger.Enter(r.Context(), "organizationsgt.RemoveMember", &request)
	defer logger.Exit(ctx, "organizationsgt.RemoveMember", &res, &err)

	if request.AccountId == "" {
		http.Error(w, "account_id is required", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	if err = accountorganizations.RemoveMember(ctx, organizationId, request.AccountId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = cascadeFullDeleteIfNoMembershipsRemain(ctx, request.AccountId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// cascadeFullDeleteIfNoMembershipsRemain soft-deletes accountId once it has
// zero remaining AccountOrganization rows (this really was its last
// facility, not just one of several). Only Parent accounts are ever
// touched here -- Director/Teacher/CoDirector accounts are single-org by
// design and RemoveMember is a parent-specific operation.
func cascadeFullDeleteIfNoMembershipsRemain(ctx context.Context, accountId string) error {
	remaining, err := accountorganizations.Select(ctx, accountorganizationspg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: accountId}}},
			},
		},
	})
	if err != nil {
		return err
	}
	if len(remaining.AccountOrganizations) > 0 {
		return nil
	}

	accountResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: accountId}}}},
		},
	})
	if err != nil {
		return err
	}
	if len(accountResult.Accounts) == 0 || accountResult.Accounts[0].AccountType != accountspg.AccountTypeParent {
		return nil
	}

	_, err = accounts.Delete(ctx, accountspg.DeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: accountId}}},
			},
		},
	})
	return err
}
