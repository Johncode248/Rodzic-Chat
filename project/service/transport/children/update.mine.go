package childrengt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func errStatus(err error) int {
	if err == children.ErrNotYourChild {
		return http.StatusForbidden
	}
	if err == children.ErrChildNotFound {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// UpdateMine handles PATCH /private/children/mine?id= -- any parent linked
// to the child may edit its name/birth date/notes/PESEL (see
// children.UpdateMine's doc comment).
func UpdateMine(w http.ResponseWriter, r *http.Request) {
	var (
		req childrenpg.UpdateMineRequest
		res childrenpg.Child
		err error
	)

	ctx := logger.Enter(r.Context(), "childrengt.UpdateMine", r.Body)
	defer logger.Exit(ctx, "childrengt.UpdateMine", &res, &err)

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = children.UpdateMine(ctx, accountId, id, req)
	if err != nil {
		http.Error(w, err.Error(), errStatus(err))
		return
	}

	json.NewEncoder(w).Encode(res)
}

// RevealPesel handles GET /private/children/mine/pesel?id= -- the "eye
// icon" endpoint, the only place a child's PESEL is ever decrypted.
func RevealPesel(w http.ResponseWriter, r *http.Request) {
	var (
		res childrenpg.RevealPeselResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "childrengt.RevealPesel")
	defer logger.Exit(ctx, "childrengt.RevealPesel", &res, &err)

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

	res, err = children.RevealPesel(ctx, accountId, id)
	if err != nil {
		http.Error(w, err.Error(), errStatus(err))
		return
	}

	json.NewEncoder(w).Encode(res)
}
