package parentchildgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errCannotAssignOtherParent = errors.New("a parent may only link themselves to a child, not another account")

func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request parentchildpg.CreateRequest
		res     parentchildpg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "parentchildgt.Create", r.Body)
	defer logger.Exit(ctx, "parentchildgt.Create", &res, &err)

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

	accountType, err := credentials.GetAccountTypeFromContext(r.Context())
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	// A parent (non-staff) may only ever create a link for THEMSELVES --
	// without this, any authenticated parent could grant an arbitrary
	// account (including a stranger's) guardian access over any child in
	// the organization, which is a child-safety issue, not just a privacy
	// one. Staff (director/co-director/teacher) retain full discretion to
	// link any parent to any child in their org, per parentInOrganization's
	// doc comment below.
	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff && request.ParentId != accountId {
		http.Error(w, errCannotAssignOtherParent.Error(), http.StatusForbidden)
		return
	}

	inOrg, err := childInOrganization(ctx, request.ChildId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, errChildNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	parentOk, err := parentInOrganization(ctx, request.ParentId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !parentOk {
		http.Error(w, errParentNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	res, err = parentchild.Create(ctx, parentchildpg.CreateArgument{
		ParentId: request.ParentId,
		ChildId:  request.ChildId,
		Relation: parentchildpg.MapRelationType(request.Relation),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
