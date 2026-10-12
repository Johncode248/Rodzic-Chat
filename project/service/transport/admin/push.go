package admingt

import (
	"context"
	"net/http"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

type audience struct {
	// Mode: "accounts" (explicit ids), "organization", "type" (account_type ordinal) or "all".
	Mode           string   `json:"mode"`
	AccountIds     []string `json:"account_ids"`
	OrganizationId string   `json:"organization_id"`
	AccountType    int      `json:"account_type"`
	// Search matches name/email; used by the UI's user picker (mode "search").
	Search string `json:"search"`
}

type recipient struct {
	Id, Name, Email string
	Devices         int
}

func resolveAudience(ctx context.Context, a audience) ([]recipient, error) {
	base := `SELECT a.id, a.name, a.email, (SELECT count(*) FROM device_tokens d WHERE d.account_id = a.id) AS devices
 FROM accounts a WHERE a.deleted_at IS NULL`
	var (
		out  []recipient
		args []any
		sql  = base
	)
	switch a.Mode {
	case "accounts":
		if len(a.AccountIds) == 0 {
			return nil, nil
		}
		sql += " AND a.id IN ?"
		args = append(args, a.AccountIds)
	case "organization":
		sql += " AND a.organization_id = ?"
		args = append(args, a.OrganizationId)
	case "type":
		sql += " AND a.account_type = ?"
		args = append(args, a.AccountType)
	case "search":
		s := "%" + strings.TrimSpace(a.Search) + "%"
		sql += " AND (a.name ILIKE ? OR a.email ILIKE ?) LIMIT 20"
		args = append(args, s, s)
	case "all":
	default:
		return nil, nil
	}
	err := gormdbScan(ctx, &out, sql, args...)
	return out, err
}

// pushAudience previews who a push would reach (accounts + how many have a registered device).
func pushAudience(w http.ResponseWriter, r *http.Request) {
	var a audience
	if !readBody(w, r, &a) {
		return
	}
	rec, err := resolveAudience(r.Context(), a)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	type item struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Devices int    `json:"devices"`
	}
	items := make([]item, 0, len(rec))
	devices := 0
	for _, x := range rec {
		items = append(items, item{x.Id, x.Name, x.Email, x.Devices})
		devices += x.Devices
	}
	writeJSON(w, 200, map[string]any{
		"accounts": len(items), "devices": devices, "sample": firstN(items, 20),
		"fcm_enabled": push.Enabled(r.Context()),
	})
}

func firstN[T any](s []T, n int) []T {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// pushSend: POST {audience, title, body, data}.
func pushSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Audience audience          `json:"audience"`
		Title    string            `json:"title"`
		Body     string            `json:"body"`
		Data     map[string]string `json:"data"`
	}
	if !readBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Body) == "" {
		fail(w, 400, "tytuł i treść są wymagane")
		return
	}
	rec, err := resolveAudience(r.Context(), req.Audience)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	ids := make([]string, 0, len(rec))
	for _, x := range rec {
		ids = append(ids, x.Id)
	}
	if req.Data == nil {
		req.Data = map[string]string{}
	}
	if _, ok := req.Data["type"]; !ok {
		req.Data["type"] = "admin"
	}
	rep := push.SendToAccountsReport(r.Context(), ids, req.Title, req.Body, req.Data)
	writeJSON(w, 200, map[string]any{"accounts": len(ids), "report": rep})
}
