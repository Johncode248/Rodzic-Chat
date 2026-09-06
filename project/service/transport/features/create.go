package featuresgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	featurespg "github.com/Johncode248/Rodzic-Chat/project/common/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/features"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Create adds a single OrganizationFeature row for the caller's org (mainly
// useful to fill in a feature added after an organization was already
// created, since register.go seeds every known feature up front).
func Create(w http.ResponseWriter, r *http.Request) {
	var (
		request featurespg.CreateRequest
		res     featurespg.CreateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "featuresgt.Create", r.Body)
	defer logger.Exit(ctx, "featuresgt.Create", &res, &err)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	organizationId, err := credentials.GetOrganizationIdFromContext(r.Context())
	if err != nil || organizationId == "" {
		http.Error(w, "Error getting organizationId from context", http.StatusForbidden)
		return
	}

	res, err = features.Create(ctx, featurespg.CreateArgument{
		OrganizationId: organizationId,
		Feature:        request.Feature,
		Enabled:        request.Enabled,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
