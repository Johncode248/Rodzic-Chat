package absencesgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/absences"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// Select lists absences in the caller's organization. Staff see every
// absence; a parent only ever sees absences for their own children.
// Optional ?date_from=/?date_to= (YYYY-MM-DD or RFC3339) filter to
// absences overlapping that range -- e.g. a director's calendar view
// clicking a single day passes the same date as both.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res absencespg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "absencesgt.Select")
	defer logger.Exit(ctx, "absencesgt.Select", &res, &err)

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

	query = query.AddOrganizationAuth(organizationId)
	query.OrderBy = databasepg.Order{Key: "date_from", Direction: databasepg.DirectionDescending}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = absences.Select(ctx, absencespg.SelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
