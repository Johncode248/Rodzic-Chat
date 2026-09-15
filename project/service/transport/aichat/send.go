package aichatgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	aichatpg "github.com/Johncode248/Rodzic-Chat/project/common/aichat"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/aichat"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentsubscriptions"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Send handles POST /private/ai-chat/messages -- any authenticated account
// (mobile-only client-side; see aichat.Send's doc comment for the full
// pipeline: plan/limit check, Anthropic call, usage bookkeeping).
func Send(w http.ResponseWriter, r *http.Request) {
	var (
		request aichatpg.SendRequest
		res     aichatpg.SendResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "aichatgt.Send", r.Body)
	defer logger.Exit(ctx, "aichatgt.Send", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = aichat.Send(ctx, accountId, request)
	if err != nil {
		switch {
		case errors.Is(err, parentsubscriptions.ErrAiLimitReached):
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		case errors.Is(err, aichat.ErrNotConfigured):
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(res)
}
