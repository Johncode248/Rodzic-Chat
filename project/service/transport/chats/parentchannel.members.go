package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectParentChannelMembers handles GET
// /private/chats/parent-channels/members?channel_id= -- the roster for a
// channel's settings screen. Caller must be a member.
func SelectParentChannelMembers(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.ChannelMemberListResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectParentChannelMembers")
	defer logger.Exit(ctx, "chatsgt.SelectParentChannelMembers", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	channelId := r.URL.Query().Get("channel_id")
	if channelId == "" {
		http.Error(w, "Missing 'channel_id' query param", http.StatusBadRequest)
		return
	}

	isMember, err := chats.IsChannelMember(ctx, channelId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "caller is not a member of this channel", http.StatusForbidden)
		return
	}

	res, err = chats.ListMembers(ctx, channelId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// RemoveParentChannelMember handles DELETE
// /private/chats/parent-channels/members?channel_id=&account_id= -- both
// "opuść czat" (account_id == caller) and "usuń członka" (admin removing
// someone else); see chats.RemoveMember's doc comment for the exact rules.
func RemoveParentChannelMember(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := logger.Enter(r.Context(), "chatsgt.RemoveParentChannelMember")
	defer logger.Exit(ctx, "chatsgt.RemoveParentChannelMember", &err)

	callerAccountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || callerAccountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	channelId := r.URL.Query().Get("channel_id")
	targetAccountId := r.URL.Query().Get("account_id")
	if channelId == "" || targetAccountId == "" {
		http.Error(w, "Missing 'channel_id' or 'account_id' query param", http.StatusBadRequest)
		return
	}

	if err = chats.RemoveMember(ctx, channelId, targetAccountId, callerAccountId); err != nil {
		status := http.StatusInternalServerError
		if err == chats.ErrNotChannelAdmin || err == chats.ErrCannotRemoveAdmin {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
}
