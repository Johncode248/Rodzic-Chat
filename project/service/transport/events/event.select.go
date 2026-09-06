package eventsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectEvents lists events in the caller's organization, open to any
// authenticated org member, optionally filtered by ?group_id=.
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

	query := databasepg.Query{
		Groups: []databasepg.Group{{}},
	}

	if groupId := r.URL.Query().Get("group_id"); groupId != "" {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "group_id", Operator: "=", Value: groupId,
		})
	}

	query = query.AddOrganizationAuth(organizationId)

	res, err = events.SelectEvents(ctx, eventspg.EventSelectArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
