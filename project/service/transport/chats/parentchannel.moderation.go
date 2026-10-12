package chatsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// requireMember writes the right 4xx/5xx and returns false unless accountId may use channelId:
// a banned member gets the ban's own explanatory message instead of a generic "not a member".
func requireMember(w http.ResponseWriter, r *http.Request, channelId, accountId string) bool {
	ctx := r.Context()
	isMember, err := chats.IsChannelMember(ctx, channelId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	if isMember {
		return true
	}
	if banErr := chats.CheckNotBanned(ctx, channelId, accountId); banErr != nil {
		http.Error(w, banErr.Error(), http.StatusForbidden)
		return false
	}
	http.Error(w, "caller is not a member of this channel", http.StatusForbidden)
	return false
}

func moderationStatus(err error) int {
	switch {
	case errors.Is(err, chats.ErrNotChannelAdmin), errors.Is(err, chats.ErrNoPermission),
		errors.Is(err, chats.ErrCannotModerateTarget), errors.Is(err, chats.ErrCannotRemoveAdmin):
		return http.StatusForbidden
	case errors.Is(err, chats.ErrNotAMember):
		return http.StatusNotFound
	case errors.Is(err, chats.ErrRoleNotAssignable), errors.Is(err, chats.ErrBanTooLong):
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// ChannelRoles handles GET /private/chats/roles -- the static catalog of chat roles and
// permissions with their descriptions (one source of truth for both mobile clients).
func ChannelRoles(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(chatspg.ChannelRolesCatalog())
}

// SetParentChannelMemberRole handles PATCH /private/chats/parent-channels/members (admin only).
func SetParentChannelMemberRole(w http.ResponseWriter, r *http.Request) {
	var (
		req chatspg.SetMemberRoleRequest
		err error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.SetParentChannelMemberRole", r.Body)
	defer logger.Exit(ctx, "chatsgt.SetParentChannelMemberRole", &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChannelId == "" || req.AccountId == "" {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	caller, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || caller == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	if err = chats.SetMemberRole(ctx, req, caller); err != nil {
		http.Error(w, err.Error(), moderationStatus(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// BanParentChannelMember handles POST /private/chats/parent-channels/bans.
func BanParentChannelMember(w http.ResponseWriter, r *http.Request) {
	var (
		req chatspg.BanRequest
		err error
	)
	ctx := logger.Enter(r.Context(), "chatsgt.BanParentChannelMember", r.Body)
	defer logger.Exit(ctx, "chatsgt.BanParentChannelMember", &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChannelId == "" || req.AccountId == "" {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}
	caller, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || caller == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	if err = chats.Ban(ctx, req, caller); err != nil {
		http.Error(w, err.Error(), moderationStatus(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UnbanParentChannelMember handles DELETE /private/chats/parent-channels/bans?channel_id=&account_id=.
func UnbanParentChannelMember(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := logger.Enter(r.Context(), "chatsgt.UnbanParentChannelMember")
	defer logger.Exit(ctx, "chatsgt.UnbanParentChannelMember", &err)

	caller, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || caller == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	channelId, accountId := r.URL.Query().Get("channel_id"), r.URL.Query().Get("account_id")
	if channelId == "" || accountId == "" {
		http.Error(w, "Missing 'channel_id' or 'account_id' query param", http.StatusBadRequest)
		return
	}
	if err = chats.Unban(ctx, channelId, accountId, caller); err != nil {
		http.Error(w, err.Error(), moderationStatus(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteParentChannelMessage handles DELETE /private/chats/parent-channels/messages?message_id= --
// the author, or a member holding PermDeleteMessages who outranks the author.
func DeleteParentChannelMessage(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := logger.Enter(r.Context(), "chatsgt.DeleteParentChannelMessage")
	defer logger.Exit(ctx, "chatsgt.DeleteParentChannelMessage", &err)

	caller, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || caller == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	messageId := r.URL.Query().Get("message_id")
	if messageId == "" {
		http.Error(w, "Missing 'message_id' query param", http.StatusBadRequest)
		return
	}
	message, err := chats.LoadMessage(ctx, messageId)
	if err != nil {
		status := http.StatusInternalServerError
		if err == chats.ErrMessageNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	if !requireMember(w, r, message.ChannelId, caller) {
		return
	}
	if err = chats.DeleteParentChannelMessage(ctx, message, caller); err != nil {
		http.Error(w, err.Error(), moderationStatus(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
