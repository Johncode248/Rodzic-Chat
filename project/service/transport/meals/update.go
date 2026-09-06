package mealsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Update edits a meal cancellation identified by ?id=. Same ownership rule
// as Delete: staff may edit any cancellation in their organization, a
// parent only one they reported themselves.
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request mealspg.UpdateRequest
		res     mealspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "mealsgt.Update", r.Body)
	defer logger.Exit(ctx, "mealsgt.Update", &res, &err)

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

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: id},
				},
			},
		},
	}
	query = query.AddOrganizationAuth(organizationId)

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "reported_by", Operator: "=", Value: accountId,
		})
	}

	res, err = meals.Update(ctx, mealspg.UpdateArgument{
		Query:    query,
		DateFrom: request.DateFrom,
		DateTo:   request.DateTo,
		Reason:   request.Reason,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.MealCancellation.Id == "" {
		http.Error(w, errNotCancellationOwner.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(res)
}
