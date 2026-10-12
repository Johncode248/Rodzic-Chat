package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// PinMessage handles POST /private/chats/messages/pin for an organization channel. Routed behind
// the staff-only middleware (see module.go): pinning is a moderation action, parents don't do it
// here. The message must belong to the caller's organization.
func PinMessage(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessagePinRequest
		res     chatspg.MessagePinResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.PinMessage", r.Body)
	defer logger.Exit(ctx, "chatsgt.PinMessage", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.MessageId == "" {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

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

	message, err := chats.LoadMessage(ctx, request.MessageId)
	if err != nil {
		status := http.StatusInternalServerError
		if err == chats.ErrMessageNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	if message.OrganizationId != organizationId {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	res.Message, err = chats.SetMessagePinned(ctx, request.MessageId, accountId, request.Pinned)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// PinParentChannelMessage handles POST /private/chats/parent-channels/messages/pin -- the org-less
// counterpart: needs the pin permission (admins and moderators have it; the admin can grant it to others).
func PinParentChannelMessage(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessagePinRequest
		res     chatspg.MessagePinResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.PinParentChannelMessage", r.Body)
	defer logger.Exit(ctx, "chatsgt.PinParentChannelMessage", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.MessageId == "" {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	message, err := chats.LoadMessage(ctx, request.MessageId)
	if err != nil {
		status := http.StatusInternalServerError
		if err == chats.ErrMessageNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	if !requireMember(w, r, message.ChannelId, accountId) {
		return
	}
	canPin, err := chats.HasChannelPermission(ctx, message.ChannelId, accountId, chatspg.PermPin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !canPin {
		http.Error(w, chats.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	res.Message, err = chats.SetMessagePinned(ctx, request.MessageId, accountId, request.Pinned)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
