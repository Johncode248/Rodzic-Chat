package parentinvitesgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Create handles POST /private/parent-invites -- "rodzic nr 1" generates a
// shareable link/token for "rodzic nr 2".
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		res parentinvitespg.CreateResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "parentinvitesgt.Create")
	defer logger.Exit(ctx, "parentinvitesgt.Create", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	invite, plaintextPin, err := parentinvites.Create(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = parentinvitespg.CreateResponse{
		Pin:       plaintextPin,
		ExpiresAt: invite.ExpiresAt.Format("2006-01-02"),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
