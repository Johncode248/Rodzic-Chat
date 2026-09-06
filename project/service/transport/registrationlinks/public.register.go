package registrationlinksgt

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
	registrationlinkspg "github.com/Johncode248/Rodzic-Chat/project/common/registrationlinks"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/registrationlinks"
	"github.com/gorilla/mux"
)

// PublicRegister is unauthenticated — completing it is how a parent gets
// their very first account. Every validation error (expired link, name
// already used in this group, email already taken, child already has two
// parents, ...) is returned as a plain 4xx with a human-readable message,
// since this is rendered straight into a public web form.
func PublicRegister(w http.ResponseWriter, r *http.Request) {
	var (
		request registrationlinkspg.PublicRegisterRequest
		res     registrationlinkspg.PublicRegisterResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "registrationlinksgt.PublicRegister", r.Body)
	defer logger.Exit(ctx, "registrationlinksgt.PublicRegister", &res, &err)

	token := mux.Vars(r)["token"]
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
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

	res, err = registrationlinks.PublicRegister(ctx, registrationlinkspg.PublicRegisterArgument{
		Token:          token,
		ChildFirstName: request.ChildFirstName,
		ChildLastName:  request.ChildLastName,
		ChildBirthDate: childBirthDate,
		Name:           request.Name,
		Email:          request.Email,
		Password:       request.Password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
