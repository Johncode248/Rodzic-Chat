package aichatgt

import (
	"net/http"
)

// Send handles POST /private/ai-chat/messages -- hard-disabled server-side for MVP, regardless
// of whether any client UI hides it (the mobile "Asystent AI" tab is already pulled from view,
// but that alone would only hide the button; a direct API call would still reach
// project/service/internal/aichat.Send otherwise). project/common/parentsubscriptions was
// stripped down to just Plan for MVP -- there is no usage/quota tracking left to protect this
// endpoint's real Anthropic API cost, so it's blocked here entirely instead. Remove this block
// (and re-wire to aichat.Send) once real usage limits are wired up.
func Send(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "AI assistant is currently disabled", http.StatusServiceUnavailable)
}
