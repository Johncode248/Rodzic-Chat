package calendargt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	calendarpg "github.com/Johncode248/Rodzic-Chat/project/common/calendar"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/calendar"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Get handles GET /private/calendar -- see calendar.Get's doc comment.
func Get(w http.ResponseWriter, r *http.Request) {
	var (
		res calendarpg.GetResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "calendargt.Get")
	defer logger.Exit(ctx, "calendargt.Get", &res, &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = calendar.Get(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
