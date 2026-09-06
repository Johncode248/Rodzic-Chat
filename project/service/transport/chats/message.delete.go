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

// DeleteMessage removes a message identified by ?id=. Only the sender or
// staff (director/teacher) may delete it.
func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.MessageDeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.DeleteMessage")
	defer logger.Exit(ctx, "chatsgt.DeleteMessage", &res, &err)

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

	accountType, err := credentials.GetAccountTypeFromContext(ctx)
	if err != nil {
		http.Error(w, "Error getting accountType from context", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing 'id' query param", http.StatusBadRequest)
		return
	}

	query := databasepg.Query{
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}}},
	}
	query = query.AddOrganizationAuth(organizationId)

	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if !isStaff {
		query.Groups[0].Filters = append(query.Groups[0].Filters, databasepg.Filter{
			Key: "sender_id", Operator: "=", Value: accountId,
		})
	}

	res, err = chats.DeleteMessage(ctx, chatspg.MessageDeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(res.Messages) == 0 {
		http.Error(w, "message not found or not owned by caller", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(res)
}
