package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// UpsertReaction adds (or replaces) the caller's reaction on a message.
func UpsertReaction(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessageReactionCreateRequest
		res     chatspg.MessageReactionCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.UpsertReaction", r.Body)
	defer logger.Exit(ctx, "chatsgt.UpsertReaction", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	inOrg, err := messageInOrganization(ctx, request.MessageId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	res, err = chats.UpsertReaction(ctx, chatspg.MessageReactionCreateArgument{
		MessageId: request.MessageId,
		AccountId: accountId,
		Emoji:     request.Emoji,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
