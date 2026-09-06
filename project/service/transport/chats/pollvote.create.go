package chatsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// CastVote casts or changes the caller's vote on a poll.
func CastVote(w http.ResponseWriter, r *http.Request) {
	var (
		request chatspg.PollVoteCreateRequest
		res     chatspg.PollVoteCreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "chatsgt.CastVote", r.Body)
	defer logger.Exit(ctx, "chatsgt.CastVote", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = chats.CreateVote(ctx, chatspg.PollVoteCreateArgument{
		PollId:    request.PollId,
		OptionId:  request.OptionId,
		AccountId: accountId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
