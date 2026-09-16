package eventsgt

import (
	"encoding/json"
	"net/http"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// DeclareRegistration is a parent's "będzie / nie będzie" declaration for
// their child on an event (or staff declaring on a parent's behalf). It is
// an upsert: repeated declarations for the same (event, child) update the
// existing row instead of creating a duplicate, enforced in
// events.Declare. Rejected once the event's registration deadline passed.
func DeclareRegistration(w http.ResponseWriter, r *http.Request) {
	var (
		request eventspg.RegistrationCreateRequest
		res     eventspg.RegistrationCreateResponse
		err     error
	)

	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	_, inOrg, err := eventInOrganization(ctx, request.EventId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, errEventNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	inOrgChild, err := childInOrganization(ctx, request.ChildId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrgChild {
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

	res, err = events.Declare(ctx, eventspg.RegistrationCreateArgument{
		EventId:      request.EventId,
		ChildId:      request.ChildId,
		ConfirmedBy:  accountId,
		Status:       request.Status,
		ConsentGiven: request.ConsentGiven,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if err == events.ErrRegistrationDeadlinePassed || err == events.ErrConsentRequired {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
