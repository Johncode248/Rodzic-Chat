package eventsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// CreateEvent creates an event or trip. Staff-only (enforced by the staffOnly
// middleware in module.go); organizationId and createdBy are taken from the
// JWT, never from the client body.
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	var (
		request eventspg.EventCreateRequest
		res     eventspg.EventCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "eventsgt.CreateEvent", r.Body)
	defer logger.Exit(ctx, "eventsgt.CreateEvent", &res, &err)

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

	if request.GroupId != "" {
		// nothing further to check here: group ownership is enforced only
		// loosely (groups themselves are org-scoped elsewhere); an event's
		// GroupId is treated as informational scoping.
	}

	res, err = events.CreateEvent(ctx, eventspg.EventCreateArgument{
		OrganizationId:       organizationId,
		GroupId:              request.GroupId,
		Type:                 request.Type,
		Title:                request.Title,
		Description:          request.Description,
		RequiresConsent:      request.RequiresConsent,
		RequiresItem:         request.RequiresItem,
		EventDate:            request.EventDate,
		RegistrationDeadline: request.RegistrationDeadline,
		CreatedBy:            accountId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
