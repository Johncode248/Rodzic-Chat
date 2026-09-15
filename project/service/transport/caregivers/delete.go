package caregiversgt

import (
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/logger"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Delete handles DELETE /private/caregivers?type= -- removes the caller's
// own profile for that type.
func Delete(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := logger.Enter(r.Context(), "caregiversgt.Delete")
	defer logger.Exit(ctx, "caregiversgt.Delete", &err)

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting accountId from context", http.StatusForbidden)
		return
	}

	typeOrdinal, convErr := strconv.Atoi(r.URL.Query().Get("type"))
	if convErr != nil {
		http.Error(w, "Missing or invalid 'type' query param", http.StatusBadRequest)
		return
	}

	if err = caregivers.Delete(ctx, accountId, caregiverspg.CaregiverTypeFromOrdinal(typeOrdinal)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
