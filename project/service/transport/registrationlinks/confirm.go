package registrationlinksgt

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Confirm handles POST /private/registration-invites/confirm -- an
// already-authenticated parent attaches to a child using a PIN. Rate
// limited (see module.go). Every validation error (expired/exhausted
// invite, child already has two parents, ...) is returned as a plain 4xx
// with a human-readable message.
func Confirm(w http.ResponseWriter, r *http.Request) {
	var (
		request registrationlinkspg.ConfirmRequest
		res     registrationlinkspg.ConfirmResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "registrationlinksgt.Confirm", r.Body)
	defer logger.Exit(ctx, "registrationlinksgt.Confirm", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	var childBirthDate *time.Time
	if request.ChildBirthDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", request.ChildBirthDate)
		if parseErr != nil {
			http.Error(w, "child_birth_date must be in YYYY-MM-DD format", http.StatusBadRequest)
			return
		}
		childBirthDate = &parsed
	}

	res, err = registrationlinks.Confirm(ctx, registrationlinkspg.ConfirmArgument{
		Pin:            request.Pin,
		AccountId:      accountId,
		ChildId:        request.ChildId,
		ChildFirstName: request.ChildFirstName,
		ChildLastName:  request.ChildLastName,
		ChildBirthDate: childBirthDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
