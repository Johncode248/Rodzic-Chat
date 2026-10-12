package coparentsgt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/coparents"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// List handles GET /private/co-parents.
func List(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "coparentsgt.List")
	var err error
	defer logger.Exit(ctx, "coparentsgt.List", &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	res, err := coparents.List(ctx, accountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// Disconnect handles DELETE /private/co-parents?parent_id=.
func Disconnect(w http.ResponseWriter, r *http.Request) {
	ctx := logger.Enter(r.Context(), "coparentsgt.Disconnect")
	var err error
	defer logger.Exit(ctx, "coparentsgt.Disconnect", &err)

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}
	parentId := r.URL.Query().Get("parent_id")
	if parentId == "" {
		http.Error(w, "parent_id is required", http.StatusBadRequest)
		return
	}
	res, err := coparents.Disconnect(ctx, accountId, parentId)
	switch {
	case errors.Is(err, coparents.ErrNotACoParent):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, coparents.ErrSelf):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}
