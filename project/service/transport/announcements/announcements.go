package announcementsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/announcements"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// List handles GET /private/announcements.
func List(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "announcementsgt.List")
	var err error
	defer logger.Exit(ctx, "announcementsgt.List", &err)
	if _, err = credentials.GetUserIdFromContext(r.Context()); err != nil {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	res, err := announcements.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}
