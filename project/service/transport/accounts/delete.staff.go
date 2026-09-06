package accountsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errTargetNotStaffInOrg = errors.New("target account is not a teacher or co-director in this organization")

// DeleteStaff lets a director/co-director delete a teacher or co-director
// account within their own organization (identified by ?id=) -- distinct
// from the self-service Delete above, which only ever targets the caller's
// own account. A director account can never be removed through this route
// (only via self-delete), avoiding an accidental "lock everyone out of the
// org" mistake.
func DeleteStaff(w http.ResponseWriter, r *http.Request) {
	var (
		res accountspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "accountsgt.DeleteStaff")
	defer logger.Exit(ctx, "accountsgt.DeleteStaff", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	targetId := r.URL.Query().Get("id")
	if targetId == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	targetResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: targetId},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(targetResult.Accounts) == 0 {
		http.Error(w, "account not found in this organization", http.StatusNotFound)
		return
	}

	target := targetResult.Accounts[0]
	if target.AccountType != accountspg.AccountTypeTeacher && target.AccountType != accountspg.AccountTypeCoDirector {
		http.Error(w, errTargetNotStaffInOrg.Error(), http.StatusForbidden)
		return
	}

	// accounts.Delete already soft-deletes the Account, hard-deletes its
	// Credential, and revokes every Session (see
	// project/service/internal/accounts/delete.go) -- same machinery
	// self-delete uses.
	res, err = accounts.Delete(ctx, accountspg.DeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: targetId}}},
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
