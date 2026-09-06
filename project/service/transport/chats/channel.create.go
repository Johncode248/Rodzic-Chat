package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// CreateChannel creates a channel. Staff-only (enforced by the staffOnly
// middleware in module.go).
func CreateChannel(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.ChannelCreateRequest
		res     chatspg.ChannelCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.CreateChannel", r.Body)
	defer logger.Exit(ctx, "chatsgt.CreateChannel", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = chats.CreateChannel(ctx, chatspg.ChannelCreateArgument{
		OrganizationId: organizationId,
		GroupId:        request.GroupId,
		Type:           request.Type,
		Name:           request.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
