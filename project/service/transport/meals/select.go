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
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists meal cancellations in the caller's organization. Staff see
// every cancellation; a parent only ever sees cancellations for their own
// children. Optional ?date_from=/?date_to= filter to cancellations
// overlapping that range. Optional ?q= searches by child first/last name
// (two-step lookup, see childIdsMatchingQuery -- MealCancellation has no
// denormalized name of its own).
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res mealspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "mealsgt.Select")
	defer logger.Exit(ctx, "mealsgt.Select", &res, &err)

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

	query := databasepg.Query{
		Groups: []databasepg.Group{{}},
	}

	if childId := r.URL.Query().Get("child_id"); childId != "" {
		inOrg, err := childInOrganization(ctx, childId, organizationId)
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
			owns, err := parentOwnsChild(ctx, accountId, childId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !owns {
				http.Error(w, errNotChildsParent.Error(), http.StatusForbidden)
				return
			}
		}

		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "child_id", Operator: "=", Value: childId,
		})
	} else {
		isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
		if !isStaff {
			childIds, err := childIdsForParent(ctx, accountId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
				Key: "child_id", Operator: "IN", Value: childIds,
			})
		}
	}

	query.Groups[0].Filters = append(query.Groups[0].Filters, dateRangeFilters(r.URL.Query().Get("date_from"), r.URL.Query().Get("date_to"))...)

	if q := r.URL.Query().Get("q"); q != "" {
		matchingChildIds, err := childIdsMatchingQuery(ctx, organizationId, q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "child_id", Operator: "IN", Value: matchingChildIds,
		})
	}

	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = meals.Select(ctx, mealspg.SelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
