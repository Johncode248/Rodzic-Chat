package mealsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

var errNotCancellationOwner = errors.New("only the reporting parent or staff may modify this meal cancellation")

// Delete removes a meal cancellation identified by ?id=. Staff may delete
// any cancellation in their organization; a parent may only delete one they
// reported themselves.
func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		res mealspg.DeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "mealsgt.Delete")
	defer logger.Exit(ctx, "mealsgt.Delete", &res, &err)

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

	res, err = meals.Delete(ctx, mealspg.DeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(res.MealCancellations) == 0 {
		http.Error(w, errNotCancellationOwner.Error(), http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(res)
}
