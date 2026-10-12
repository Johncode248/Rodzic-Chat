package admingt

import (
	"fmt"
	"net/http"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/gorilla/mux"
)

// accountOverview: the account row plus, for every table that points at it, how many rows it owns.
func accountOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	acc, err := getTable(ctx, "accounts")
	if err != nil {
		fail(w, 500, err.Error())
		return
	}
	rows := []map[string]any{}
	gormdb.GormDB.WithContext(ctx).Raw("SELECT "+joinCols(visibleCols(acc))+" FROM accounts WHERE id = ?", id).Scan(&rows)
	if len(rows) == 0 {
		fail(w, 404, "nie ma takiego konta")
		return
	}
	normalize(rows)

	type rel struct {
		Table  string `json:"table"`
		Column string `json:"column"`
		Count  int64  `json:"count"`
	}
	related := []rel{}
	schemaMu.Lock()
	var cands []rel
	for _, n := range tableNames {
		if n == "accounts" {
			continue
		}
		for _, c := range schemaCache[n].Columns {
			switch c.Name {
			case "account_id", "sender_id", "uploader_id", "parent_id", "author_id", "created_by", "user_id":
				cands = append(cands, rel{Table: n, Column: c.Name})
			}
		}
	}
	schemaMu.Unlock()
	for _, c := range cands {
		var n int64
		sql := fmt.Sprintf("SELECT count(*) FROM %s WHERE %s::text = ?", q(c.Table), q(c.Column))
		if gormdb.GormDB.WithContext(ctx).Raw(sql, id).Scan(&n).Error == nil && n > 0 {
			c.Count = n
			related = append(related, c)
		}
	}
	writeJSON(w, 200, map[string]any{"account": rows[0], "related": related})
}

func joinCols(c []string) string {
	s := ""
	for i, v := range c {
		if i > 0 {
			s += ", "
		}
		s += v
	}
	return s
}

// setAccountStatus: 1 enabled, 2 disabled, 3 banned (accountspg.Status).
func setAccountStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status int `json:"status"`
	}
	if !readBody(w, r, &body) {
		return
	}
	if body.Status < 1 || body.Status > 3 {
		fail(w, 400, "status musi być 1, 2 lub 3")
		return
	}
	res := gormdb.GormDB.WithContext(r.Context()).Exec("UPDATE accounts SET status = ?, update_time = now() WHERE id = ?", body.Status, mux.Vars(r)["id"])
	if res.Error != nil {
		fail(w, 400, res.Error.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "affected": res.RowsAffected})
}
