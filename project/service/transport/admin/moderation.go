package admingt

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

// A moderation source is one kind of user-generated content. Sources whose table/columns are
// missing from the database are skipped, so this list can safely name things ahead of the schema.
type source struct {
	Key    string   `json:"key"`
	Label  string   `json:"label"`
	Table  string   `json:"-"`
	Text   string   `json:"-"` // SQL expression for the visible text
	Author string   `json:"-"` // column holding the author account id
	Time   string   `json:"-"`
	Extra  string   `json:"-"` // SQL expression for a secondary line (channel, category...)
	Needs  []string `json:"-"`
}

var sources = []source{
	{Key: "messages", Label: "Wiadomości czatu", Table: "messages", Text: "coalesce(nullif(t.content,''), t.attachment_url, '')", Author: "sender_id", Time: "create_time", Extra: "'kanał ' || t.channel_id", Needs: []string{"content", "sender_id", "create_time", "channel_id", "attachment_url"}},
	{Key: "help_requests", Label: "Prośby „Szukam / Polecam”", Table: "help_requests", Text: "t.title || ' — ' || coalesce(t.description,'')", Author: "account_id", Time: "create_time", Extra: "t.category || ' · ' || coalesce(t.city,'')", Needs: []string{"title", "description", "account_id", "create_time", "category", "city"}},
	{Key: "help_replies", Label: "Odpowiedzi na prośby", Table: "help_replies", Text: "t.text", Author: "account_id", Time: "create_time", Extra: "'prośba ' || t.request_id", Needs: []string{"text", "account_id", "create_time", "request_id"}},
	{Key: "media", Label: "Galeria (zdjęcia)", Table: "media", Text: "coalesce(nullif(t.caption,''), t.original_filename, '')", Author: "uploader_id", Time: "create_time", Extra: "t.content_type", Needs: []string{"caption", "original_filename", "uploader_id", "create_time", "content_type"}},
	{Key: "caregivers", Label: "Profile opiekunek / korepetytorów", Table: "caregivers", Text: "t.display_name || ' — ' || coalesce(t.description,'')", Author: "account_id", Time: "create_time", Extra: "t.city", Needs: []string{"display_name", "description", "account_id", "create_time", "city"}},
}

func availableSources(ctx context.Context) []source {
	var out []source
	for _, s := range sources {
		t, err := getTable(ctx, s.Table)
		if err != nil {
			continue
		}
		ok := true
		for _, c := range s.Needs {
			if t.col(c) == nil {
				ok = false
			}
		}
		if ok && len(t.PK) == 1 {
			out = append(out, s)
		}
	}
	return out
}

func moderationSources(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"sources": availableSources(r.Context())})
}

func findSource(ctx context.Context, key string) *source {
	for _, s := range availableSources(ctx) {
		if s.Key == key {
			return &s
		}
	}
	return nil
}

// moderationFeed: GET /admin/api/moderation?source=messages&q=&page=1 -- newest content first,
// with the author resolved so it can be removed / the author banned straight from the row.
func moderationFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s := findSource(ctx, r.URL.Query().Get("source"))
	if s == nil {
		fail(w, 404, "nieznane źródło treści")
		return
	}
	t, _ := getTable(ctx, s.Table)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const per = 40
	where, args := "", []any{}
	if qs := strings.TrimSpace(r.URL.Query().Get("q")); qs != "" {
		where = " WHERE (" + s.Text + ") ILIKE ? OR a.name ILIKE ? OR a.email ILIKE ?"
		args = append(args, "%"+qs+"%", "%"+qs+"%", "%"+qs+"%")
	}
	from := fmt.Sprintf(" FROM %s t LEFT JOIN accounts a ON a.id::text = t.%s::text", q(s.Table), q(s.Author))
	var total int64
	gormdb.GormDB.WithContext(ctx).Raw("SELECT count(*)"+from+where, args...).Scan(&total)
	rows := []map[string]any{}
	sql := fmt.Sprintf(`SELECT t.%s::text AS id, (%s) AS text, (%s) AS extra, t.%s AS created,
 t.%s::text AS author_id, a.name AS author_name, a.email AS author_email, a.status AS author_status%s%s
 ORDER BY t.%s DESC LIMIT %d OFFSET %d`,
		q(t.PK[0]), s.Text, s.Extra, q(s.Time), q(s.Author), from, where, q(s.Time), per, (page-1)*per)
	if err := gormdb.GormDB.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		fail(w, 400, err.Error())
		return
	}
	normalize(rows)
	writeJSON(w, 200, map[string]any{"items": rows, "total": total, "page": page, "per_page": per})
}

// moderationRemove: POST {source, id, reason, notify, ban_author, delete_author}.
// Removes the content; optionally tells the author why by push and bans/deletes the account.
func moderationRemove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Source       string `json:"source"`
		Id           string `json:"id"`
		Reason       string `json:"reason"`
		Notify       bool   `json:"notify"`
		BanAuthor    bool   `json:"ban_author"`
		DeleteAuthor bool   `json:"delete_author"`
	}
	if !readBody(w, r, &req) {
		return
	}
	s := findSource(ctx, req.Source)
	if s == nil || req.Id == "" {
		fail(w, 400, "nieznane źródło lub brak id")
		return
	}
	t, _ := getTable(ctx, s.Table)
	var authors []string
	gormdb.GormDB.WithContext(ctx).Raw(fmt.Sprintf("SELECT %s::text FROM %s WHERE %s::text = ?", q(s.Author), q(s.Table), q(t.PK[0])), req.Id).Scan(&authors)
	if len(authors) == 0 {
		fail(w, 404, "treść już nie istnieje")
		return
	}
	author := authors[0]
	if _, err := hardDelete(ctx, t, map[string]any{t.PK[0]: req.Id}); err != nil {
		fail(w, 400, err.Error())
		return
	}
	// Notify before a possible account deletion (device tokens go with the account).
	if req.Notify && author != "" {
		reason := req.Reason
		if reason == "" {
			reason = "naruszenie regulaminu"
		}
		push.SendToAccountsReport(ctx, []string{author}, "Twoja treść została usunięta", "Powód: "+reason, map[string]string{"type": "moderation"})
	}
	if req.DeleteAuthor && author != "" {
		gormdb.GormDB.WithContext(ctx).Exec("DELETE FROM accounts WHERE id = ?", author)
	} else if req.BanAuthor && author != "" {
		gormdb.GormDB.WithContext(ctx).Exec("UPDATE accounts SET status = 3, update_time = now() WHERE id = ?", author)
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
