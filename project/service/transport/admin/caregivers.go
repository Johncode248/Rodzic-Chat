package admingt

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Johncode248/Rodzic-Chat/project/service/internal/caregivers"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

// caregiversList: GET /admin/api/caregivers?status=1|2|3 (pending / approved / rejected).
func caregiversList(w http.ResponseWriter, r *http.Request) {
	status, err := strconv.Atoi(r.URL.Query().Get("status"))
	if err != nil || status < 1 || status > 3 {
		status = 1
	}
	items, err := caregivers.ListForReview(r.Context(), status)
	if err != nil {
		fail(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

// caregiversDecide: POST {id, approve, reason}; the owner is told by push about the outcome.
func caregiversDecide(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Id      string `json:"id"`
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}
	if !readBody(w, r, &req) || req.Id == "" {
		if req.Id == "" {
			fail(w, 400, "brak id")
		}
		return
	}
	// Look up the owner first so the push can go out after a successful decision.
	type owner struct{ AccountId, DisplayName string }
	var o owner
	rows := []owner{}
	_ = queryInto(r, &rows, "SELECT account_id, display_name FROM caregivers WHERE id = ?", req.Id)
	if len(rows) > 0 {
		o = rows[0]
	}
	if err := caregivers.Decide(r.Context(), req.Id, req.Approve, req.Reason); err != nil {
		code := http.StatusInternalServerError
		switch {
		case errors.Is(err, caregivers.ErrNotFound):
			code = http.StatusNotFound
		case errors.Is(err, caregivers.ErrReasonRequired):
			code = http.StatusBadRequest
		}
		fail(w, code, err.Error())
		return
	}
	if o.AccountId != "" {
		title, body := "Twój profil został zatwierdzony", "Jest już widoczny w katalogu."
		if !req.Approve {
			title, body = "Profil wymaga poprawek", req.Reason
		}
		go push.SendToAccounts(context.Background(), []string{o.AccountId}, title, body, map[string]string{"type": "caregiver_review"})
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
