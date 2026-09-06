package mealsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errChildNotInOrganization = errors.New("child does not belong to the caller's organization")
var errNotChildsParent = errors.New("caller is not a registered parent of this child")

// Create reports a meal cancellation ("odwołanie obiadu" -- the child is
// present but will not eat the provided meal on the given date(s)). Staff
// (director/teacher) may report one for any child in their organization; a
// parent may only report one for a child they are actually linked to via
// ParentChild.
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request mealspg.CreateRequest
		res     mealspg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "mealsgt.Create", r.Body)
	defer logger.Exit(ctx, "mealsgt.Create", &res, &err)

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

	res, err = meals.Create(ctx, mealspg.CreateArgument{
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
