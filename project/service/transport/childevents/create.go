package childeventsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Create handles POST /private/child-events -- any parent linked to the
// child (see project/service/internal/parentchild.IsParent) can add a
// personal calendar entry for them.
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		req childeventspg.CreateRequest
		res childeventspg.ChildEvent
		err error
	)

	ctx := logger.Enter(r.Context(), "childeventsgt.Create", r.Body)
	defer logger.Exit(ctx, "childeventsgt.Create", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	res, err = childevents.Create(ctx, accountId, req)
	if err != nil {
		if err == childevents.ErrNotYourChild {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
