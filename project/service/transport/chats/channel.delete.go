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

// DeleteChannel removes a channel identified by ?id=. Staff-only.
func DeleteChannel(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.ChannelDeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.DeleteChannel")
	defer logger.Exit(ctx, "chatsgt.DeleteChannel", &res, &err)

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
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

	res, err = chats.DeleteChannel(ctx, chatspg.ChannelDeleteArgument{Query: query})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
