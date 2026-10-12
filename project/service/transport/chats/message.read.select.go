package chatsgt

import (
	"encoding/json"
	"net/http"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// SelectReads lists who has read ?message_id= -- the "who viewed this
// announcement" feature. Staff see the full roster; a parent can only see
// whether THEY read it (results are filtered to their own AccountId).
func SelectReads(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageReadSelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.SelectReads")
	defer logger.Exit(ctx, "chatsgt.SelectReads", &res, &err)

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

	messageId := r.URL.Query().Get("message_id")
	if messageId == "" {
		http.Error(w, "Missing 'message_id' query param", http.StatusBadRequest)
		return
	}

	// Note: organizationId may legitimately be "" here -- see
	// messageAccessible.
	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	channel, found, allowed, err := messageAccessible(ctx, messageId, organizationId, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}
	if !allowed {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}
	if err = chats.CheckReadPermission(ctx, channel, accountId, accountType); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	filters := []databasepg.Filter{
		{Key: "message_id", Operator: "=", Value: messageId},
	}

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff {
		filters = append(filters, databasepg.Filter{Key: "account_id", Operator: "=", Value: accountId})
	}

	res, err = chats.SelectMessageReads(ctx, chatspg.MessageReadSelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: filters}}},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
