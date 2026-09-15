package chatsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// MarkParentChannelMessageRead handles POST /private/chats/parent-channels/messages/reads --
// the org-less counterpart of MarkRead, authorized via chatspg.ChannelMember instead of
// organization_id (see parentchannel.select.go's doc comment on why org-less routes exist at
// all). Body: {"message_id": "..."}.
func MarkParentChannelMessageRead(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessageReadCreateRequest
		res     chatspg.MessageReadCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.MarkParentChannelMessageRead", r.Body)
	defer logger.Exit(ctx, "chatsgt.MarkParentChannelMessageRead", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err = requireParentChannelMessageMember(ctx, request.MessageId, accountId); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	res, err = chats.MarkRead(ctx, chatspg.MessageReadCreateArgument{
		MessageId: request.MessageId,
		AccountId: accountId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// SelectParentChannelMessageReads handles GET
// /private/chats/parent-channels/messages/reads?message_id= -- the org-less counterpart of
// SelectReads. Unlike the staff/parent split in SelectReads (announcements gate who may see the
// full roster), any member of a parent-to-parent channel may see who else has read a message --
// there's no staff/parent distinction in a peer chat.
func SelectParentChannelMessageReads(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageReadSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectParentChannelMessageReads")
	defer logger.Exit(ctx, "chatsgt.SelectParentChannelMessageReads", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	messageId := r.URL.Query().Get("message_id")
	if messageId == "" {
		http.Error(w, "Missing 'message_id' query param", http.StatusBadRequest)
		return
	}

	if err = requireParentChannelMessageMember(ctx, messageId, accountId); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	res, err = chats.SelectMessageReads(ctx, chatspg.MessageReadSelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "message_id", Operator: "=", Value: messageId}}}}},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
