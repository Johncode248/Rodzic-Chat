package helprequests

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	hp "github.com/Johncode248/Rodzic-Chat/project/common/helprequests"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("request not found")
	ErrForbidden      = errors.New("not allowed")
	ErrTitle          = errors.New("title is required (max 120 characters)")
	ErrCategory       = errors.New("unknown category")
	ErrText           = errors.New("recommendation text is required (max 1000 characters)")
	ErrLink           = errors.New("link must start with http:// or https://")
	ErrClosed         = errors.New("this request is closed")
	ErrTooManyReplies = errors.New("you already gave the maximum number of recommendations here")
	ErrRating         = errors.New("rating must be between 1 and 5")
	ErrPhoto          = errors.New("unknown or foreign photo")
)

func byId(id string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}}}}
}

// shortName: first name plus the initial of the last one -- what other users see.
func shortName(full string) string {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "Użytkownik"
	case 1:
		return parts[0]
	}
	return parts[0] + " " + string([]rune(parts[len(parts)-1])[:1]) + "."
}

type listRow struct {
	hp.HelpRequest
	AuthorName        string
	ReplyCount        int64
	PreviewText       string
	PreviewAuthorName string
	UnseenReplies     int64
}

// ListFilter drives GET /private/requests. Scope: "" (open feed), "unanswered", "solved", "mine", "nearby" (needs City).
// A non-empty Q searches the whole catalog (open AND solved questions). Sort: "" (newest) or "popular".
type ListFilter struct {
	Q, Category, City, Scope, Sort string
	Limit, Offset                  int
}

// listSelect is shared by List and Detail: the request plus author, reply count, an answer preview and the author's unseen replies.
const listSelect = `select r.*, a.name as author_name,
   (select count(*) from help_replies x where x.request_id = r.id) as reply_count,
   coalesce(pr.text, '') as preview_text, coalesce(pa.name, '') as preview_author_name,
   (select count(*) from help_replies x where x.request_id = r.id and x.account_id <> r.account_id
      and x.create_time > coalesce(r.owner_seen_time, r.create_time)) as unseen_replies
 from help_requests r join accounts a on a.id = r.account_id
 left join lateral (select x.* from help_replies x where x.request_id = r.id
   order by (x.id = r.best_reply_id) desc, (select count(*) from help_reply_votes v where v.reply_id = x.id) desc, x.create_time asc limit 1) pr on true
 left join accounts pa on pa.id = pr.account_id`

func toItem(r listRow, viewerId string) hp.RequestItem {
	item := hp.RequestItem{HelpRequest: r.HelpRequest, AuthorName: shortName(r.AuthorName), ReplyCount: r.ReplyCount, IsMine: r.AccountId == viewerId,
		PreviewText: clip(r.PreviewText, 160), PreviewAuthor: shortName(r.PreviewAuthorName)}
	if r.PreviewText == "" {
		item.PreviewAuthor = ""
	}
	if item.IsMine {
		item.UnseenReplies = r.UnseenReplies
	}
	return item
}

func List(ctx context.Context, viewerId string, f ListFilter) (hp.ListResponse, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	q := strings.TrimSpace(f.Q)
	switch f.Scope {
	case "mine":
		where = append(where, "r.account_id = ?")
		args = append(args, viewerId)
	case "solved":
		where = append(where, "r.solved = true")
	case "unanswered":
		where = append(where, "r.closed = false and r.solved = false and not exists (select 1 from help_replies x where x.request_id = r.id)")
	default: // open feed ("nearby" is the open feed narrowed by city)
		if q != "" {
			where = append(where, "(r.solved = true or r.closed = false)")
		} else {
			where = append(where, "r.closed = false and r.solved = false")
		}
	}
	if q != "" {
		like := "%" + q + "%"
		where = append(where, "(r.title ilike ? or r.description ilike ?)")
		args = append(args, like, like)
	}
	if f.Category != "" {
		where = append(where, "r.category = ?")
		args = append(args, f.Category)
	}
	if city := strings.TrimSpace(f.City); city != "" {
		where = append(where, "r.city ilike ?")
		args = append(args, "%"+city+"%")
	}
	order := "r.create_time desc"
	if f.Sort == "popular" {
		order = "reply_count desc, r.create_time desc"
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	args = append(args, limit, f.Offset)
	rows, err := gormdb.CustomSelect[listRow](ctx, listSelect+" where "+strings.Join(where, " and ")+" order by "+order+" limit ? offset ?", args...)
	if err != nil {
		return hp.ListResponse{}, err
	}
	out := make([]hp.RequestItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, toItem(r, viewerId))
	}
	return hp.ListResponse{Requests: out}, nil
}

func Create(ctx context.Context, accountId string, req hp.CreateRequest) (hp.HelpRequest, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" || len([]rune(title)) > 120 {
		return hp.HelpRequest{}, ErrTitle
	}
	ok := false
	for _, c := range hp.Categories {
		if c == req.Category {
			ok = true
		}
	}
	if !ok {
		return hp.HelpRequest{}, ErrCategory
	}
	if err := checkPhoto(ctx, accountId, req.PhotoId); err != nil {
		return hp.HelpRequest{}, err
	}
	now := time.Now()
	r := hp.HelpRequest{Id: uuid.NewString(), AccountId: accountId, Title: title, Category: req.Category, PhotoId: req.PhotoId,
		Description: clip(strings.TrimSpace(req.Description), 1000), City: clip(strings.TrimSpace(req.City), 80), OwnerSeenTime: now, CreateTime: now, UpdateTime: now}
	if err := gormdb.Create(ctx, &r); err != nil {
		return r, err
	}
	go notifyFollowers(context.WithoutCancel(ctx), r)
	return r, nil
}

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

func get(ctx context.Context, id string) (*hp.HelpRequest, error) {
	rows, err := gormdb.Select[hp.HelpRequest](ctx, byId(id))
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

func Detail(ctx context.Context, viewerId, id string) (hp.DetailResponse, error) {
	list, err := gormdb.CustomSelect[listRow](ctx, listSelect+" where r.id = ?", id)
	if err != nil {
		return hp.DetailResponse{}, err
	}
	if len(list) == 0 {
		return hp.DetailResponse{}, ErrNotFound
	}
	r := list[0]
	type replyRow struct {
		hp.HelpReply
		AuthorName    string
		CaregiverName string
		CaregiverType int
		VoteCount     int64
		VotedByMe     bool
	}
	replies, err := gormdb.CustomSelect[replyRow](ctx, `select x.*, a.name as author_name, c.display_name as caregiver_name, coalesce(c.type,0) as caregiver_type,
   (select count(*) from help_reply_votes v where v.reply_id = x.id) as vote_count,
   exists(select 1 from help_reply_votes v where v.reply_id = x.id and v.account_id = ?) as voted_by_me
 from help_replies x join accounts a on a.id = x.account_id left join caregivers c on c.id = x.caregiver_id and x.caregiver_id <> ''
 where x.request_id = ? order by (x.id = ?) desc, vote_count desc, x.create_time asc`, viewerId, id, r.BestReplyId)
	if err != nil {
		return hp.DetailResponse{}, err
	}
	item := toItem(r, viewerId)
	out := hp.DetailResponse{Request: item, Replies: []hp.ReplyItem{}}
	for _, x := range replies {
		out.Replies = append(out.Replies, hp.ReplyItem{HelpReply: x.HelpReply, AuthorName: shortName(x.AuthorName), IsMine: x.AccountId == viewerId,
			CaregiverName: x.CaregiverName, CaregiverType: x.CaregiverType, VoteCount: x.VoteCount, VotedByMe: x.VotedByMe,
			IsBest: r.BestReplyId != "" && x.Id == r.BestReplyId})
	}
	if item.IsMine {
		// Opening your own question marks its replies as seen (clears the bell entry).
		gormdb.GormDB.WithContext(ctx).Exec("update help_requests set owner_seen_time = ? where id = ?", time.Now(), id)
	}
	return out, nil
}

func SetClosed(ctx context.Context, viewerId, id string, closed bool) error {
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
	r.Closed, r.UpdateTime = closed, time.Now()
	return gormdb.Update(ctx, *r, byId(id))
}

func Delete(ctx context.Context, viewerId, id string) error {
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
	_, err = gormdb.Delete[hp.HelpRequest](ctx, byId(id))
	return err
}

func Reply(ctx context.Context, viewerId string, req hp.ReplyRequest) (hp.HelpReply, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" || len([]rune(text)) > 1000 {
		return hp.HelpReply{}, ErrText
	}
	link := strings.TrimSpace(req.Url)
	if link != "" {
		if u, err := url.Parse(link); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return hp.HelpReply{}, ErrLink
		}
	}
	r, err := get(ctx, req.RequestId)
	if err != nil {
		return hp.HelpReply{}, err
	}
	if r == nil {
		return hp.HelpReply{}, ErrNotFound
	}
	if r.Closed {
		return hp.HelpReply{}, ErrClosed
	}
	mine, err := gormdb.CustomSelect[struct{ N int }](ctx, "select count(*) as n from help_replies where request_id = ? and account_id = ?", req.RequestId, viewerId)
	if err != nil {
		return hp.HelpReply{}, err
	}
	if len(mine) > 0 && mine[0].N >= 5 {
		return hp.HelpReply{}, ErrTooManyReplies
	}
	if req.Rating < 0 || req.Rating > 5 {
		return hp.HelpReply{}, ErrRating
	}
	if err := checkPhoto(ctx, viewerId, req.PhotoId); err != nil {
		return hp.HelpReply{}, err
	}
	reply := hp.HelpReply{Id: uuid.NewString(), RequestId: req.RequestId, AccountId: viewerId, Text: text, Url: link, CaregiverId: req.CaregiverId,
		PhotoId: req.PhotoId, ProductName: clip(strings.TrimSpace(req.ProductName), 80), Price: clip(strings.TrimSpace(req.Price), 40),
		Rating: req.Rating, UsedSince: clip(strings.TrimSpace(req.UsedSince), 40), CreateTime: time.Now()}
	if err := gormdb.Create(ctx, &reply); err != nil {
		return reply, err
	}
	if r.AccountId != viewerId {
		name := reply.ProductName
		if name == "" {
			name = "Nowe polecenie"
		}
		go push.SendToAccounts(context.WithoutCancel(ctx), []string{r.AccountId}, "Ktoś odpowiedział na Twoje pytanie", name+" – "+clip(r.Title, 80),
			map[string]string{"type": "help_reply", "request_id": r.Id})
	}
	return reply, nil
}

// DeleteReply: the reply's author or the request's author may remove it.
func DeleteReply(ctx context.Context, viewerId, id string) error {
	rows, err := gormdb.Select[hp.HelpReply](ctx, byId(id))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return ErrNotFound
	}
	req, err := get(ctx, rows[0].RequestId)
	if err != nil {
		return err
	}
	if rows[0].AccountId != viewerId && (req == nil || req.AccountId != viewerId) {
		return ErrForbidden
	}
	_, err = gormdb.Delete[hp.HelpReply](ctx, byId(id))
	return err
}

var ErrOwnReply = errors.New("you cannot back your own recommendation")

// ToggleVote adds or removes the caller's "Też polecam" on somebody else's reply; returns whether it is now set.
func ToggleVote(ctx context.Context, viewerId, replyId string) (bool, error) {
	rows, err := gormdb.Select[hp.HelpReply](ctx, byId(replyId))
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, ErrNotFound
	}
	if rows[0].AccountId == viewerId {
		return false, ErrOwnReply
	}
	req, err := get(ctx, rows[0].RequestId)
	if err != nil {
		return false, err
	}
	if req != nil && req.Closed {
		return false, ErrClosed
	}
	q := databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{
		{Key: "reply_id", Operator: "=", Value: replyId}, {Key: "account_id", Operator: "=", Value: viewerId}}}}}
	existing, err := gormdb.Select[hp.HelpReplyVote](ctx, q)
	if err != nil {
		return false, err
	}
	if len(existing) > 0 {
		_, err = gormdb.Delete[hp.HelpReplyVote](ctx, q)
		return false, err
	}
	return true, gormdb.Create(ctx, &hp.HelpReplyVote{ReplyId: replyId, AccountId: viewerId, CreateTime: time.Now()})
}
