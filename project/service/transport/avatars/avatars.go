package avatarsgt

import (
	"io"
	"net/http"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	avatarspg "github.com/Johncode248/Rodzic-Chat/project/common/avatars"
	"github.com/Johncode248/Rodzic-Chat/project/service/transport/credentials"
)

func byAccount(id string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "account_id", Operator: "=", Value: id}}}}}
}

// canSee: yourself, or someone in the same organization, or someone you share a chat channel with.
func canSee(r *http.Request, viewerId, targetId string) bool {
	if viewerId == targetId {
		return true
	}
	type row struct{ N int }
	rows, err := gormdb.CustomSelect[row](r.Context(), `
select count(*) as n from (
  select 1 from accounts a join accounts b on a.organization_id = b.organization_id and a.organization_id <> ''
   where a.id = ? and b.id = ? and a.deleted_at is null and b.deleted_at is null
  union all
  select 1 from caregivers where account_id = ? and (review_status = 2 or review_status = 0)
  union all
  select 1 from channel_members m1 join channel_members m2 on m1.channel_id = m2.channel_id
   where m1.account_id = ? and m2.account_id = ?
) t`, viewerId, targetId, targetId, viewerId, targetId)
	return err == nil && len(rows) > 0 && rows[0].N > 0
}

// Get handles GET /private/accounts/avatar?id= (defaults to the caller): the raw image, or 404.
func Get(w http.ResponseWriter, r *http.Request) {
	viewerId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || viewerId == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	target := r.URL.Query().Get("id")
	if target == "" {
		target = viewerId
	}
	if !canSee(r, viewerId, target) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	list, err := gormdb.Select[avatarspg.Avatar](r.Context(), byAccount(target))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", list[0].ContentType)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.Write(list[0].Data)
}

// Put handles PUT /private/accounts/avatar: the body is the image itself (JPEG or PNG).
func Put(w http.ResponseWriter, r *http.Request) {
	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, avatarspg.MaxBytes+1))
	if err != nil {
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}
	if len(data) == 0 || len(data) > avatarspg.MaxBytes {
		http.Error(w, "Image must be between 1 byte and 512 KB", http.StatusBadRequest)
		return
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "Only JPEG or PNG images are accepted", http.StatusUnsupportedMediaType)
		return
	}
	// Replace: delete any previous picture, then insert.
	if _, err = gormdb.Delete[avatarspg.Avatar](r.Context(), byAccount(accountId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = gormdb.Create(r.Context(), &avatarspg.Avatar{AccountId: accountId, ContentType: contentType, Data: data, UpdateTime: time.Now()}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /private/accounts/avatar.
func Delete(w http.ResponseWriter, r *http.Request) {
	accountId, err := credentials.GetUserIdFromContext(r.Context())
	if err != nil || accountId == "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if _, err = gormdb.Delete[avatarspg.Avatar](r.Context(), byAccount(accountId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
