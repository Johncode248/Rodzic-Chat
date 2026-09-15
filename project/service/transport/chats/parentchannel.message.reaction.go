package chatsgt

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// requireParentChannelMessageMember looks up which channel a message belongs to and confirms
// accountId is a member of it -- the shared authorization step for every org-less reaction/read
// route in this file and in parentchannel.message.read.go (see that file's doc comment on why
// these routes exist at all: the org-scoped UpsertReaction/SelectReactions/DeleteReaction are
// wrapped in chatFeature, which 403s unconditionally for an org-less caller since there's no
// organizationId to check).
func requireParentChannelMessageMember(ctx context.Context, messageId, accountId string) error {
	messages, err := chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit:  1,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: messageId}}}},
		},
	})
	if err != nil {
		return err
	}
	if len(messages.Messages) == 0 {
		return errors.New("message not found")
	}

	isMember, err := chats.IsChannelMember(ctx, messages.Messages[0].ChannelId, accountId)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("caller is not a member of this channel")
	}
	return nil
}

// UpsertParentChannelMessageReaction handles POST /private/chats/parent-channels/messages/reactions.
func UpsertParentChannelMessageReaction(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.MessageReactionCreateRequest
		res     chatspg.MessageReactionCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.UpsertParentChannelMessageReaction", r.Body)
	defer logger.Exit(ctx, "chatsgt.UpsertParentChannelMessageReaction", &res, &err)

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

// SelectParentChannelMessageReactions handles GET
// /private/chats/parent-channels/messages/reactions?message_id=.
func SelectParentChannelMessageReactions(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageReactionSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectParentChannelMessageReactions")
	defer logger.Exit(ctx, "chatsgt.SelectParentChannelMessageReactions", &res, &err)

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

	res, err = chats.SelectReactions(ctx, chatspg.MessageReactionSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "message_id", Operator: "=", Value: messageId}}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// DeleteParentChannelMessageReaction handles DELETE
// /private/chats/parent-channels/messages/reactions?message_id=.
func DeleteParentChannelMessageReaction(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageReactionDeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.DeleteParentChannelMessageReaction")
	defer logger.Exit(ctx, "chatsgt.DeleteParentChannelMessageReaction", &res, &err)

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

	res, err = chats.DeleteReaction(ctx, chatspg.MessageReactionDeleteArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "message_id", Operator: "=", Value: messageId},
				{Key: "account_id", Operator: "=", Value: accountId},
			}}},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
