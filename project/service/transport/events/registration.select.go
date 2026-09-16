package eventsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/pagination"
)

// SelectRegistrations lists registrations, org-scoped via the event.
// Filterable by ?event_id= or ?child_id=. Staff see every registration in
// the org; a parent only ever sees registrations for their own children.
func SelectRegistrations(w http.ResponseWriter, r *http.Request) {
	var (
		res eventspg.RegistrationSelectResponse
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

	filters := []databasepg.Filter{}

	if eventId := r.URL.Query().Get("event_id"); eventId != "" {
		_, inOrg, err := eventInOrganization(ctx, eventId, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !inOrg {
			http.Error(w, errEventNotInOrganization.Error(), http.StatusForbidden)
			return
		}
		filters = append(filters, databasepg.Filter{Key: "event_id", Operator: "=", Value: eventId})
	} else {
		eventIds, err := eventIdsInOrganization(ctx, organizationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filters = append(filters, databasepg.Filter{Key: "event_id", Operator: "IN", Value: eventIds})
	}

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()

	if childId := r.URL.Query().Get("child_id"); childId != "" {
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
		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "=", Value: childId})
	} else if !isStaff {
		childIds, err := childIdsForParent(ctx, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filters = append(filters, databasepg.Filter{Key: "child_id", Operator: "IN", Value: childIds})
	}

	query := databasepg.Query{
		Groups:  []databasepg.Group{{Filters: filters}},
		OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
	}
	query.Limit, query.Offset = pagination.Parse(r, pagination.DefaultMaxLimit)

	res, err = events.SelectRegistrations(ctx, eventspg.RegistrationSelectArgument{
		Query: query,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
