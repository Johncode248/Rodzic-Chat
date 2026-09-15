package childrengt

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// CreateMine handles POST /private/children/mine -- a parent adding a
// child entirely on their own (no placowka involved). See
// children.CreateMine's doc comment.
func CreateMine(w http.ResponseWriter, r *http.Request) {
	var (
		request childrenpg.CreateMineRequest
		res     childrenpg.CreateMineResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "childrengt.CreateMine", r.Body)
	defer logger.Exit(ctx, "childrengt.CreateMine", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	var birthDate time.Time
	if request.BirthDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", request.BirthDate)
		if parseErr != nil {
			http.Error(w, "birth_date must be in YYYY-MM-DD format", http.StatusBadRequest)
			return
		}
		birthDate = parsed
	}

	res, err = children.CreateMine(ctx, childrenpg.CreateMineArgument{
		ParentAccountId: accountId,
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		BirthDate:       birthDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// SelectMine handles GET /private/children/mine -- every child the caller
// is linked to as a parent, personal or placowka-connected alike. This is
// the "moje dzieci" list the parent-facing mobile navigation is built
// around, distinct from childrengt.Select's org-scoped staff roster.
func SelectMine(w http.ResponseWriter, r *http.Request) {
	var (
		res childrenpg.SelectMineResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "childrengt.SelectMine")
	defer logger.Exit(ctx, "childrengt.SelectMine", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = children.SelectMine(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
