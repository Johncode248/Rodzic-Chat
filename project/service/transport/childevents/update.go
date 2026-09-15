package childeventsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Update handles PATCH /private/child-events?id= -- partial update, same
// authorization as Create (any linked parent may edit).
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		req childeventspg.UpdateRequest
		res childeventspg.ChildEvent
		err error
	)

	ctx := logger.Enter(r.Context(), "childeventsgt.Update", r.Body)
	defer logger.Exit(ctx, "childeventsgt.Update", &res, &err)

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

	res, err = childevents.Update(ctx, accountId, id, req)
	if err != nil {
		if err == childevents.ErrNotYourChild {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if err == childevents.ErrEventNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(res)
}

// Delete handles DELETE /private/child-events?id=
func Delete(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := logger.Enter(r.Context(), "childeventsgt.Delete")
	defer logger.Exit(ctx, "childeventsgt.Delete", &err)

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

	if err = childevents.Delete(ctx, accountId, id); err != nil {
		if err == childevents.ErrNotYourChild {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if err == childevents.ErrEventNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
