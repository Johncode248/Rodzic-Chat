package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// CreateParentChannelMessage handles POST
// /private/chats/parent-channels/messages -- the org-less counterpart of
// CreateMessage, authorized via chatspg.ChannelMember instead of
// organization_id (see parentchannel.select.go's doc comment). The target
// channel is always Type Direct (see chatinvites.Create), which
// chats.CreateMessage already leaves unrestricted -- rate limiting still
// applies.
func CreateParentChannelMessage(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessageCreateRequest
		res     chatspg.MessageCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.CreateParentChannelMessage", r.Body)
	defer logger.Exit(ctx, "chatsgt.CreateParentChannelMessage", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	isMember, err := chats.IsChannelMember(ctx, request.ChannelId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "caller is not a member of this channel", http.StatusForbidden)
		return
	}

	res, err = chats.CreateMessage(ctx, chatspg.MessageCreateArgument{
		ChannelId:      request.ChannelId,
		OrganizationId: "",
		SenderId:       accountId,
		Type:           chatspg.MessageTypeText,
		Content:        request.Content,
		AttachmentUrl:  request.AttachmentUrl,
		Urgent:         request.Urgent,
		ReplyToId:      request.ReplyToId,
	}, accountType)
	if err != nil {
		status := http.StatusForbidden
		if err == chats.ErrRateLimited {
			status = http.StatusTooManyRequests
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
