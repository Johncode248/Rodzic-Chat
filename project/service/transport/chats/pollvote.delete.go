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

// RetractVote removes the caller's own vote, identified by ?poll_id= and
// ?option_id=.
func RetractVote(w http.ResponseWriter, r *http.Request) {
	var (
		res chatspg.PollVoteDeleteResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.RetractVote")
	defer logger.Exit(ctx, "chatsgt.RetractVote", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	pollId := r.URL.Query().Get("poll_id")
	optionId := r.URL.Query().Get("option_id")
	if pollId == "" {
		http.Error(w, "Missing 'poll_id' query param", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(ctx)
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	inOrg, err := pollInOrganization(ctx, pollId, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !inOrg {
		http.Error(w, chats.ErrChannelNotInOrganization.Error(), http.StatusForbidden)
		return
	}

	filters := []databasepg.Filter{
		{Key: "poll_id", Operator: "=", Value: pollId},
		{Key: "account_id", Operator: "=", Value: accountId},
	}
	if optionId != "" {
		filters = append(filters, databasepg.Filter{Key: "option_id", Operator: "=", Value: optionId})
	}

	res, err = chats.DeleteVote(ctx, chatspg.PollVoteDeleteArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: filters}}},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
