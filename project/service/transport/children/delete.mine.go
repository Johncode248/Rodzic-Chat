package childrengt

import (
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// DeleteMine handles DELETE /private/children/mine?id= -- see children.DeleteMine's doc comment
// for exactly what happens (full delete for a personal child, just unlinking the caller for a
// placówka-connected one).
func DeleteMine(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := logger.Enter(r.Context(), "childrengt.DeleteMine", r.URL.RawQuery)
	defer logger.Exit(ctx, "childrengt.DeleteMine", &err)

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	if err = children.DeleteMine(ctx, accountId, id); err != nil {
		http.Error(w, err.Error(), errStatus(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
