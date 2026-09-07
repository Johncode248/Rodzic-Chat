package platformsettingsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	platformsettingspg "github.com/Johncode248/Rodzic-Chat/project/common/platformsettings"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/platformsettings"
)

// Update changes any subset of the platform-wide fee settings (e.g. just
// PlatformFeeGrosze to raise/lower Rodzic Chat's own per-payment cut).
// Secret-gated (see helpers.go).
func Update(w http.ResponseWriter, r *http.Request) {
	var (
		request platformsettingspg.UpdateRequest
		res     platformsettingspg.UpdateResponse
		err     error
	)

	ctx := logger.Enter(r.Context(), "platformsettingsgt.Update", r.Body)
	defer logger.Exit(ctx, "platformsettingsgt.Update", &res, &err)

	if !requirePlatformAdminSecret(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	updated, err := platformsettings.Update(ctx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = platformsettingspg.UpdateResponse{PlatformSettings: updated}
	json.NewEncoder(w).Encode(res)
}
