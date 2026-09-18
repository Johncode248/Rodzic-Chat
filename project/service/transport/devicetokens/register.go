package devicetokensgt

import (
	"encoding/json"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/logger"
	devicetokenspg "github.com/Johncode248/Rodzic-Chat/project/common/devicetokens"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/devicetokens"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

// Register handles POST /private/device-tokens: the client sends its FCM
// registration token on every app launch and right after login (tokens can
// rotate, and a fresh login may be a different account on the same
// device).
func Register(w http.ResponseWriter, r *http.Request) {
	var (
		request devicetokenspg.RegisterRequest
		res     devicetokenspg.DeviceToken
		err     error
	)

	ctx := logger.Enter(r.Context(), "devicetokensgt.Register", r.Body)
	defer logger.Exit(ctx, "devicetokensgt.Register", &res, &err)

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting userId from context", http.StatusForbidden)
		return
	}

	res, err = devicetokens.Register(ctx, accountId, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"device_token": res})
}

// Unregister handles DELETE /private/device-tokens: called on logout so a
// signed-out device stops receiving pushes for the account that just left
// it. Scoped to the caller's own account -- can't be used to remove
// someone else's token.
func Unregister(w http.ResponseWriter, r *http.Request) {
	var (
		request devicetokenspg.UnregisterRequest
		err     error
	)

	ctx := logger.Enter(r.Context(), "devicetokensgt.Unregister", r.Body)
	defer logger.Exit(ctx, "devicetokensgt.Unregister", &err)

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
		return
	}

	accountId, err := credentials.GetUserIdFromContext(ctx)
	if err != nil || accountId == "" {
		http.Error(w, "Error getting userId from context", http.StatusForbidden)
		return
	}

	if err = devicetokens.Unregister(ctx, accountId, request.Token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
