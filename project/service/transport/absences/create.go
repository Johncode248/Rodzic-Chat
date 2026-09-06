package absencesgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/absences"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errChildNotInOrganization = errors.New("child does not belong to the caller's organization")
var errNotChildsParent = errors.New("caller is not a registered parent of this child")

// Create reports an absence. Staff (director/teacher) may report an absence
// for any child in their organization; a parent may only report one for a
// child they are actually linked to via ParentChild.
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request absencespg.CreateRequest
		res     absencespg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "absencesgt.Create", r.Body)
	defer logger.Exit(ctx, "absencesgt.Create", &res, &err)

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

	inOrg, err := childInOrganization(ctx, request.ChildId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, errChildNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff {
		owns, err := parentOwnsChild(ctx, accountId, request.ChildId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !owns {
			http.Error(w, errNotChildsParent.Error(), http.StatusForbidden)
			return
		}
	}

	res, err = absences.Create(ctx, absencespg.CreateArgument{
		OrganizationId: organizationId,
		ChildId:        request.ChildId,
		ReportedBy:     accountId,
		DateFrom:       request.DateFrom,
		DateTo:         request.DateTo,
		Reason:         request.Reason,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
