package platformsettingsgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	platformsettingspg "github.com/Johncode248/Rodzic-Chat/project/common/platformsettings"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/platformsettings"
)

// Select returns the current platform-wide fee settings. Secret-gated (see
// helpers.go) -- there's nothing here an organization's own director needs
// to read directly, the numbers relevant to them come back through payment
// responses instead.
func Select(w http.ResponseWriter, r *http.Request) {
	var (
		res platformsettingspg.SelectResponse
		err error
	)

	ctx := logger.Enter(r.Context(), "platformsettingsgt.Select")
	defer logger.Exit(ctx, "platformsettingsgt.Select", &res, &err)

	if !requirePlatformAdminSecret(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	settings, err := platformsettings.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res = platformsettingspg.SelectResponse{PlatformSettings: settings}
	json.NewEncoder(w).Encode(res)
}
