package helprequests

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	hp "github.com/Johncode248/Rodzic-Chat/project/common/helprequests"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
	"github.com/google/uuid"
)

// ---- photos ----

// checkPhoto: an empty id is fine; otherwise the photo must exist and belong to the caller.
func checkPhoto(ctx context.Context, accountId, photoId string) error {
	if photoId == "" {
		return nil
	}
	rows, err := gormdb.Select[hp.HelpPhoto](ctx, byId(photoId))
	if err != nil {
		return err
	}
	if len(rows) == 0 || rows[0].AccountId != accountId {
		return ErrPhoto
	}
	return nil
}

var ErrPhotoSize = errors.New("photo must be a JPEG or PNG up to 1 MB")

// SavePhoto stores an uploaded image (the request body) and returns its id.
func SavePhoto(ctx context.Context, accountId string, body io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(body, hp.PhotoMaxBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) == 0 || len(data) > hp.PhotoMaxBytes {
		return "", ErrPhotoSize
	}
	ct := http.DetectContentType(data)
	if ct != "image/jpeg" && ct != "image/png" {
		return "", ErrPhotoSize
	}
	p := hp.HelpPhoto{Id: uuid.NewString(), AccountId: accountId, ContentType: ct, Data: data, CreateTime: time.Now()}
	return p.Id, gormdb.Create(ctx, &p)
}

func GetPhoto(ctx context.Context, id string) (*hp.HelpPhoto, error) {
	rows, err := gormdb.Select[hp.HelpPhoto](ctx, byId(id))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	return &rows[0], nil
}

// ---- best answer ----

// SetBest marks (or, with replyId == "", clears) the recommendation that helped the author.
func SetBest(ctx context.Context, viewerId, id, replyId string) error {
	r, err := get(ctx, id)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrNotFound
	}
	if r.AccountId != viewerId {
		return ErrForbidden
	}
	var replyAuthor string
	if replyId != "" {
		rows, err := gormdb.Select[hp.HelpReply](ctx, byId(replyId))
		if err != nil {
			return err
		}
		if len(rows) == 0 || rows[0].RequestId != id {
			return ErrNotFound
		}
		replyAuthor = rows[0].AccountId
	}
	r.BestReplyId, r.Solved, r.UpdateTime = replyId, replyId != "", time.Now()
	if err := gormdb.Update(ctx, *r, byId(id)); err != nil {
		return err
	}
	if replyAuthor != "" && replyAuthor != viewerId {
		go push.SendToAccounts(context.WithoutCancel(ctx), []string{replyAuthor}, "Twoje polecenie pomogło!", clip(r.Title, 80),
			map[string]string{"type": "help_best", "request_id": id})
	}
	return nil
}

// ---- similar questions ----

var stop = map[string]bool{"jakie": true, "jaki": true, "jaka": true, "szukam": true, "polecam": true, "polecicie": true, "czy": true, "dla": true, "oraz": true, "bardzo": true, "który": true, "która": true, "które": true}

func words(s string) []string {
	f := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	out := []string{}
	seen := map[string]bool{}
	for _, w := range f {
		r := []rune(w)
		if len(r) < 4 || stop[w] || seen[w] {
			continue
		}
		if len(r) > 6 { // crude Polish stemming: compare on the first letters so "wózka" matches "wózek"
			w = string(r[:len(r)-2])
		} else {
			w = string(r[:len(r)-1])
		}
		seen[w] = true
		out = append(out, w)
	}
	return out
}

// Similar returns up to 5 existing (open or solved) questions sharing words with q, solved ones first.
func Similar(ctx context.Context, viewerId, q string) (hp.ListResponse, error) {
	ws := words(q)
	if len(ws) == 0 {
		return hp.ListResponse{Requests: []hp.RequestItem{}}, nil
	}
	if len(ws) > 6 {
		ws = ws[:6]
	}
	score := []string{}
	args := []interface{}{}
	for _, w := range ws {
		score = append(score, "(case when r.title ilike ? or r.description ilike ? then 1 else 0 end)")
		args = append(args, "%"+w+"%", "%"+w+"%")
	}
	sql := listSelect + " where (r.solved = true or r.closed = false) and (" + strings.Join(score, " + ") + ") > 0 order by (" +
		strings.Join(score, " + ") + ") desc, r.solved desc, r.create_time desc limit 5"
	args = append(args, args...)
	rows, err := gormdb.CustomSelect[listRow](ctx, sql, args...)
	if err != nil {
		return hp.ListResponse{}, err
	}
	out := make([]hp.RequestItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, toItem(r, viewerId))
	}
	return hp.ListResponse{Requests: out}, nil
}

// ---- followed categories ----

func Follows(ctx context.Context, accountId string) (hp.FollowsResponse, error) {
	rows, err := gormdb.CustomSelect[hp.CategoryFollow](ctx, "select * from category_follows where account_id = ? order by category", accountId)
	if err != nil {
		return hp.FollowsResponse{}, err
	}
	out := hp.FollowsResponse{Categories: []string{}}
	for _, r := range rows {
		out.Categories = append(out.Categories, r.Category)
	}
	return out, nil
}

func SetFollow(ctx context.Context, accountId, category string, on bool) error {
	known := false
	for _, c := range hp.Categories {
		if c == category {
			known = true
		}
	}
	if !known {
		return ErrCategory
	}
	db := gormdb.GormDB.WithContext(ctx)
	if !on {
		return db.Exec("delete from category_follows where account_id = ? and category = ?", accountId, category).Error
	}
	return db.Exec("insert into category_follows (account_id, category, create_time) values (?, ?, ?) on conflict do nothing", accountId, category, time.Now()).Error
}

func notifyFollowers(ctx context.Context, r hp.HelpRequest) {
	rows, err := gormdb.CustomSelect[hp.CategoryFollow](ctx, "select * from category_follows where category = ? and account_id <> ?", r.Category, r.AccountId)
	if err != nil || len(rows) == 0 {
		return
	}
	ids := make([]string, 0, len(rows))
	for _, f := range rows {
		ids = append(ids, f.AccountId)
	}
	push.SendToAccounts(ctx, ids, "Nowe pytanie: "+r.Category, clip(r.Title, 100), map[string]string{"type": "help_new", "request_id": r.Id})
}

// ---- bell activity + stats ----

func Activity(ctx context.Context, accountId string) (hp.ActivityResponse, error) {
	type row struct {
		RequestId     string
		Title         string
		UnseenReplies int64
	}
	rows, err := gormdb.CustomSelect[row](ctx, `select r.id as request_id, r.title, count(x.id) as unseen_replies
 from help_requests r join help_replies x on x.request_id = r.id and x.account_id <> r.account_id and x.create_time > coalesce(r.owner_seen_time, r.create_time)
 where r.account_id = ? group by r.id, r.title order by max(x.create_time) desc limit 20`, accountId)
	if err != nil {
		return hp.ActivityResponse{}, err
	}
	out := hp.ActivityResponse{Items: []hp.ActivityItem{}}
	for _, x := range rows {
		out.Items = append(out.Items, hp.ActivityItem{RequestId: x.RequestId, Title: x.Title, UnseenReplies: x.UnseenReplies})
	}
	return out, nil
}

func Stats(ctx context.Context, accountId string) (hp.StatsResponse, error) {
	type row struct{ Replies, BestCount, VotesReceived int64 }
	rows, err := gormdb.CustomSelect[row](ctx, `select
 (select count(*) from help_replies where account_id = ?) as replies,
 (select count(*) from help_requests r join help_replies x on x.id = r.best_reply_id where x.account_id = ? and r.account_id <> ?) as best_count,
 (select count(*) from help_reply_votes v join help_replies x on x.id = v.reply_id where x.account_id = ?) as votes_received`, accountId, accountId, accountId, accountId)
	if err != nil || len(rows) == 0 {
		return hp.StatsResponse{}, err
	}
	s := hp.StatsResponse{Replies: rows[0].Replies, BestCount: rows[0].BestCount, VotesReceived: rows[0].VotesReceived}
	s.Helped = s.BestCount + s.VotesReceived
	return s, nil
}
