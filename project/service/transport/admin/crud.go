package admingt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func tables(w http.ResponseWriter, r *http.Request) {
	if err := loadSchema(r.Context()); err != nil {
		fail(w, 500, err.Error())
		return
	}
	type row struct {
		Name    string `json:"name"`
		Count   int64  `json:"count"`
		Columns int    `json:"columns"`
	}
	type cnt struct {
		Relname string
		N       int64
	}
	// n_live_tup is an estimate, good enough for a sidebar; exact counts are computed on open.
	counts, _ := gormdb.CustomSelect[cnt](r.Context(), `select relname, n_live_tup as n from pg_stat_user_tables`)
	byName := map[string]int64{}
	for _, c := range counts {
		byName[c.Relname] = c.N
	}
	schemaMu.Lock()
	out := make([]row, 0, len(tableNames))
	for _, n := range tableNames {
		out = append(out, row{Name: n, Count: byName[n], Columns: len(schemaCache[n].Columns)})
	}
	schemaMu.Unlock()
	writeJSON(w, 200, map[string]any{"tables": out})
}

func visibleCols(t *tableInfo) []string {
	out := []string{}
	for _, c := range t.Columns {
		if !c.Hidden {
			out = append(out, q(c.Name))
		}
	}
	return out
}

func normalize(rows []map[string]any) {
	for _, row := range rows {
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
	}
}

func listRows(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t, err := getTable(ctx, mux.Vars(r)["table"])
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	qs := r.URL.Query()
	page, _ := strconv.Atoi(qs.Get("page"))
	if page < 1 {
		page = 1
	}
	per, _ := strconv.Atoi(qs.Get("per_page"))
	if per < 1 || per > 200 {
		per = 50
	}

	var where []string
	var args []any
	// f.<column>=<value> exact filters (used for "related objects" links).
	for key, vals := range qs {
		if !strings.HasPrefix(key, "f.") || len(vals) == 0 {
			continue
		}
		c := t.col(strings.TrimPrefix(key, "f."))
		if c == nil {
			continue
		}
		where = append(where, fmt.Sprintf("%s::text = ?", q(c.Name)))
		args = append(args, vals[0])
	}
	if s := strings.TrimSpace(qs.Get("q")); s != "" {
		var ors []string
		for _, c := range t.Columns {
			if c.Hidden {
				continue
			}
			ors = append(ors, fmt.Sprintf("%s::text ILIKE ?", q(c.Name)))
			args = append(args, "%"+s+"%")
		}
		if len(ors) > 0 {
			where = append(where, "("+strings.Join(ors, " OR ")+")")
		}
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	order := ""
	if c := t.col(qs.Get("sort")); c != nil && !c.Hidden {
		dir := "ASC"
		if qs.Get("dir") == "desc" {
			dir = "DESC"
		}
		order = " ORDER BY " + q(c.Name) + " " + dir
	} else if c := t.col("create_time"); c != nil {
		order = " ORDER BY " + q(c.Name) + " DESC"
	} else if len(t.PK) > 0 {
		order = " ORDER BY " + q(t.PK[0])
	}

	var total int64
	if err := gormdb.GormDB.WithContext(ctx).Raw("SELECT count(*) FROM "+q(t.Name)+whereSQL, args...).Scan(&total).Error; err != nil {
		fail(w, 400, err.Error())
		return
	}
	rows := []map[string]any{}
	sql := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT %d OFFSET %d",
		strings.Join(visibleCols(t), ", "), q(t.Name), whereSQL, order, per, (page-1)*per)
	if err := gormdb.GormDB.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		fail(w, 400, err.Error())
		return
	}
	normalize(rows)
	writeJSON(w, 200, map[string]any{"table": t, "rows": rows, "total": total, "page": page, "per_page": per})
}

// keyClause turns {"id":"x"} into `"id" = ?` for exactly the table's primary key columns.
func keyClause(t *tableInfo, key map[string]any) (string, []any, error) {
	if len(t.PK) == 0 {
		return "", nil, fmt.Errorf("tabela %s nie ma klucza głównego – edycja niemożliwa", t.Name)
	}
	var parts []string
	var args []any
	for _, pk := range t.PK {
		v, ok := key[pk]
		if !ok {
			return "", nil, fmt.Errorf("brakuje wartości klucza %q", pk)
		}
		parts = append(parts, q(pk)+"::text = ?")
		args = append(args, fmt.Sprint(v))
	}
	return strings.Join(parts, " AND "), args, nil
}

func parseKey(r *http.Request) map[string]any {
	key := map[string]any{}
	json.Unmarshal([]byte(r.URL.Query().Get("key")), &key)
	return key
}

func getRow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t, err := getTable(ctx, mux.Vars(r)["table"])
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	cl, args, err := keyClause(t, parseKey(r))
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	rows := []map[string]any{}
	if err := gormdb.GormDB.WithContext(ctx).Raw("SELECT "+strings.Join(visibleCols(t), ", ")+" FROM "+q(t.Name)+" WHERE "+cl, args...).Scan(&rows).Error; err != nil {
		fail(w, 400, err.Error())
		return
	}
	if len(rows) == 0 {
		fail(w, 404, "nie znaleziono wiersza")
		return
	}
	normalize(rows)
	writeJSON(w, 200, map[string]any{"table": t, "row": rows[0]})
}

// coerce converts a JSON value to what Postgres expects for the column's type.
func coerce(c column, v any) any {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case float64:
		if isIntType(c.Type) {
			return int64(x)
		}
		return x
	case string:
		if x == "" && c.Type != "text" && c.Type != "character varying" && c.Type != "character" {
			// blank input for a non-text column (uuid, number, date...) means NULL
			return nil
		}
		return x
	case map[string]any, []any:
		b, _ := json.Marshal(x)
		return string(b)
	}
	return v
}

func writable(t *tableInfo, values map[string]any) (cols []string, args []any) {
	for k, v := range values {
		c := t.col(k)
		if c == nil || c.Hidden {
			continue
		}
		cols = append(cols, k)
		args = append(args, coerce(*c, v))
	}
	return
}

func createRow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t, err := getTable(ctx, mux.Vars(r)["table"])
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	var body struct {
		Values map[string]any `json:"values"`
	}
	if !readBody(w, r, &body) {
		return
	}
	if body.Values == nil {
		body.Values = map[string]any{}
	}
	now := time.Now()
	for _, n := range []string{"create_time", "update_time"} {
		if t.col(n) != nil {
			if _, ok := body.Values[n]; !ok {
				body.Values[n] = now.Format(time.RFC3339Nano)
			}
		}
	}
	if len(t.PK) == 1 && t.PK[0] == "id" {
		if v, ok := body.Values["id"]; !ok || v == "" {
			body.Values["id"] = uuid.NewString()
		}
	}
	cols, args := writable(t, body.Values)
	if len(cols) == 0 {
		fail(w, 400, "brak pól do zapisania")
		return
	}
	qcols := make([]string, len(cols))
	ph := make([]string, len(cols))
	for i, c := range cols {
		qcols[i], ph[i] = q(c), "?"
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q(t.Name), strings.Join(qcols, ","), strings.Join(ph, ","))
	if err := gormdb.GormDB.WithContext(ctx).Exec(sql, args...).Error; err != nil {
		fail(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"ok": true, "id": body.Values["id"]})
}

func updateRow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t, err := getTable(ctx, mux.Vars(r)["table"])
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	var body struct {
		Key    map[string]any `json:"key"`
		Values map[string]any `json:"values"`
	}
	if !readBody(w, r, &body) {
		return
	}
	cl, kargs, err := keyClause(t, body.Key)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	if t.col("update_time") != nil {
		if _, ok := body.Values["update_time"]; !ok {
			body.Values["update_time"] = time.Now().Format(time.RFC3339Nano)
		}
	}
	cols, args := writable(t, body.Values)
	if len(cols) == 0 {
		fail(w, 400, "brak pól do zapisania")
		return
	}
	sets := make([]string, len(cols))
	for i, c := range cols {
		sets[i] = q(c) + " = ?"
	}
	res := gormdb.GormDB.WithContext(ctx).Exec("UPDATE "+q(t.Name)+" SET "+strings.Join(sets, ", ")+" WHERE "+cl, append(args, kargs...)...)
	if res.Error != nil {
		fail(w, 400, res.Error.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "affected": res.RowsAffected})
}

// hardDelete removes a row for good (FK cascades in the models do the rest).
func hardDelete(ctx context.Context, t *tableInfo, key map[string]any) (int64, error) {
	cl, args, err := keyClause(t, key)
	if err != nil {
		return 0, err
	}
	res := gormdb.GormDB.WithContext(ctx).Exec("DELETE FROM "+q(t.Name)+" WHERE "+cl, args...)
	return res.RowsAffected, res.Error
}

func deleteRow(w http.ResponseWriter, r *http.Request) {
	t, err := getTable(r.Context(), mux.Vars(r)["table"])
	if err != nil {
		fail(w, 404, err.Error())
		return
	}
	var body struct {
		Key map[string]any `json:"key"`
	}
	if !readBody(w, r, &body) {
		return
	}
	n, err := hardDelete(r.Context(), t, body.Key)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "affected": n})
}
