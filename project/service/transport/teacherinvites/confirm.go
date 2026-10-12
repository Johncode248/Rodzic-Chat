package teacherinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	teacherinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/teacherinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Confirm handles POST /private/teacher-invites/confirm -- an
// already-authenticated caller attaches to the invite's organization as
// staff using a PIN. Rate limited (see module.go).
func Confirm(w http.ResponseWriter, r *http.Request) {
	var (
		request teacherinvitespg.ConfirmRequest
		res     teacherinvitespg.ConfirmResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "teacherinvitesgt.Confirm", r.Body)
	defer logger.Exit(ctx, "teacherinvitesgt.Confirm", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	res, err = teacherinvites.Confirm(ctx, teacherinvitespg.ConfirmArgument{
		Pin:       request.Pin,
		AccountId: accountId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
