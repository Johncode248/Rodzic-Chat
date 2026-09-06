package eventsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// DeleteRegistration retracts a declaration identified by ?id=. Staff may
// delete any registration in their organization; a parent only their own
// child's.
func DeleteRegistration(w http.ResponseWriter, r *http.Request) {
	var (
		res eventspg.RegistrationDeleteResponse
		err error
	)

	ctx := r.Context()

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

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	eventIds, err := eventIdsInOrganization(ctx, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "id", Operator: "=", Value: id},
					{Key: "event_id", Operator: "IN", Value: eventIds},
				},
			},
		},
	}

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

	res, err = events.DeleteRegistration(ctx, eventspg.RegistrationDeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(res.EventRegistrations) == 0 {
		http.Error(w, "registration not found or not owned by caller", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(res)
}
