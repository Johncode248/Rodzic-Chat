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

// DeleteEvent removes an event identified by ?id=. Staff-only.
func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	var (
		res eventspg.EventDeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "eventsgt.DeleteEvent")
	defer logger.Exit(ctx, "eventsgt.DeleteEvent", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
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

	res, err = events.DeleteEvent(ctx, eventspg.EventDeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
