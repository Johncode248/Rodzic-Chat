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

var errPermissionTargetNotTeacher = errors.New("permissions can only be granted to a teacher account -- a co-director already has every permission")

// UpdateStaffPermissionsRequest carries one bool per grantable permission.
// A nil field leaves that permission untouched; new grantable permissions
// get their own field here as they're added.
type UpdateStaffPermissionsRequest struct {
	CanManagePayments *bool `json:"can_manage_payments,omitempty"`
}

// UpdateStaffPermissions lets a director/co-director grant or revoke extra
// capabilities (currently: creating/editing/deleting payments) on a teacher
// account within their own organization (identified by ?id=). Only a plain
// Teacher account can be a target -- CoDirector already has every
// permission implicitly, and this route can't be used to touch a Director.
func UpdateStaffPermissions(w http.ResponseWriter, r *http.Request) {
	var (
		request UpdateStaffPermissionsRequest
		res     accountspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "accountsgt.UpdateStaffPermissions", r.Body)
	defer logger.Exit(ctx, "accountsgt.UpdateStaffPermissions", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

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
	if target.AccountType != accountspg.AccountTypeTeacher {
		http.Error(w, errPermissionTargetNotTeacher.Error(), http.StatusForbidden)
		return
	}

	newPermissions := target.Permissions
	if request.CanManagePayments != nil {
		newPermissions = newPermissions.Set(accountspg.PermissionManagePayments, *request.CanManagePayments)
	}

	res, err = accounts.Update(ctx, accountspg.UpdateArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: targetId}}},
			},
		},
		Permissions: &newPermissions,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
