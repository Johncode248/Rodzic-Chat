package eventsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectEvents lists events in the caller's organization, optionally
// filtered by ?group_id=. Staff (director/co-director/teacher) see every
// event in the organization. A parent only sees events for their own
// children's groups, plus org-wide events (empty group_id) -- mirroring
// absencesgt.Select's staff-vs-parent scoping.
func SelectEvents(w http.ResponseWriter, r *http.Request) {
	var (
		res eventspg.EventSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "eventsgt.SelectEvents")
	defer logger.Exit(ctx, "eventsgt.SelectEvents", &res, &err)

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

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()

	var query databasepg.Query
	requestedGroupId := r.URL.Query().Get("group_id")

	if isStaff {
		if requestedGroupId != "" {
			query.Groups = []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: requestedGroupId}}},
			}
		} else {
			query.Groups = []databasepg.Group{{}}
		}
	} else {
		groupIds, err := groupIdsForParent(ctx, accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if requestedGroupId != "" {
			// A parent may only ever request a group one of their own
			// children belongs to -- silently narrows to "no results"
			// rather than leaking events from a group they have no
			// business seeing.
			allowed := false
			for _, id := range groupIds {
				if id == requestedGroupId {
					allowed = true
					break
				}
			}
			if !allowed {
				json.NewEncoder(w).Encode(eventspg.EventSelectResponse{})
				return
			}
			query.Groups = []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: requestedGroupId}}},
			}
		} else {
			// Org-wide events (no specific group) are always visible;
			// events tied to a group are visible only if one of the
			// parent's children is in that group.
			query.Groups = []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: ""}}},
			}
			if len(groupIds) > 0 {
				query.Groups = append(query.Groups, databasepg.Group{
					Filters: []databasepg.Filter{{Key: "group_id", Operator: "IN", Value: groupIds}},
				})
			}
		}
	}

	query = query.AddOrganizationAuth(organizationId)

	res, err = events.SelectEvents(ctx, eventspg.EventSelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
