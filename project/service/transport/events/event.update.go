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

// UpdateEvent edits an event identified by ?id=. Staff-only.
func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var (
		request eventspg.EventUpdateRequest
		res     eventspg.EventUpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "eventsgt.UpdateEvent", r.Body)
	defer logger.Exit(ctx, "eventsgt.UpdateEvent", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

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

	res, err = events.UpdateEvent(ctx, eventspg.EventUpdateArgument{
		Query:                query,
		Title:                request.Title,
		Description:          request.Description,
		RequiresItem:         request.RequiresItem,
		EventDate:            request.EventDate,
		RegistrationDeadline: request.RegistrationDeadline,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
