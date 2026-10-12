package admingt

import (
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

func scalar(r *http.Request, sql string) int64 {
	var n int64
	if err := gormdb.GormDB.WithContext(r.Context()).Raw(sql).Scan(&n).Error; err != nil {
		return -1 // table may not exist yet; the UI shows "–"
	}
	return n
}

func stats(w http.ResponseWriter, r *http.Request) {
	type day struct {
		Day string `json:"day"`
		N   int64  `json:"n"`
	}
	signups := []day{}
	gormdb.GormDB.WithContext(r.Context()).Raw(`SELECT to_char(d::date,'YYYY-MM-DD') AS day,
 (SELECT count(*) FROM accounts a WHERE a.create_time::date = d::date) AS n
 FROM generate_series(current_date - 13, current_date, '1 day') d ORDER BY d`).Scan(&signups)
	type bucket struct {
		AccountType int   `json:"account_type"`
		N           int64 `json:"n"`
	}
	byType := []bucket{}
	gormdb.GormDB.WithContext(r.Context()).Raw(`SELECT account_type, count(*) AS n FROM accounts WHERE deleted_at IS NULL GROUP BY account_type ORDER BY account_type`).Scan(&byType)
	writeJSON(w, 200, map[string]any{
		"accounts":           scalar(r, "SELECT count(*) FROM accounts WHERE deleted_at IS NULL"),
		"organizations":      scalar(r, "SELECT count(*) FROM organizations"),
		"children":           scalar(r, "SELECT count(*) FROM children"),
		"messages_7d":        scalar(r, "SELECT count(*) FROM messages WHERE create_time > now() - interval '7 days'"),
		"caregivers_pending": scalar(r, "SELECT count(*) FROM caregivers WHERE review_status = 1"),
		"caregivers_total":   scalar(r, "SELECT count(*) FROM caregivers"),
		"help_requests":      scalar(r, "SELECT count(*) FROM help_requests"),
		"devices":            scalar(r, "SELECT count(*) FROM device_tokens"),
		"banned":             scalar(r, "SELECT count(*) FROM accounts WHERE status = 3"),
		"fcm_enabled":        push.Enabled(r.Context()),
		"signups":            signups,
		"by_type":            byType,
	})
}
