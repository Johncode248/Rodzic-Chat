package todos

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	todospg "github.com/Johncode248/Rodzic-Chat/project/common/todos"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/dashboard"
	"github.com/google/uuid"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrTitleTooLong  = errors.New("title must be at most 200 characters")
	ErrNoteTooLong   = errors.New("note must be at most 1000 characters")
	ErrNotFound      = errors.New("todo not found")
)

func byIdAndOwner(id, accountId string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{{Filters: []databasepg.Filter{
		{Key: "id", Operator: "=", Value: id},
		{Key: "account_id", Operator: "=", Value: accountId},
	}}}}
}

// coParentIds are the accounts that share at least one child with accountId (the "drugi rodzic").
func coParentIds(ctx context.Context, accountId string) []string {
	type row struct{ ParentId string }
	rows, err := gormdb.CustomSelect[row](ctx, `select distinct pc2.parent_id from parent_children pc1
 join parent_children pc2 on pc1.child_id = pc2.child_id and pc2.parent_id <> pc1.parent_id where pc1.parent_id = ?`, accountId)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ParentId)
	}
	return ids
}

func contains(list []string, id string) bool {
	for _, x := range list {
		if x == id {
			return true
		}
	}
	return false
}

type todoRow struct {
	todospg.Todo
	OwnerName string
}

// List merges the caller's own tasks, tasks the other parent shared, and the automatic to-dos (see
// dashboard.Outstanding, minus the ones the caller snoozed): open items first ordered by due date
// (undated last), then done own tasks (newest first, max 50).
func List(ctx context.Context, accountId string) (res todospg.ListResponse, err error) {
	ctx = logger.Enter(ctx, "todos.List", accountId)
	defer logger.Exit(ctx, "todos.List", &res, &err)

	res.Items = []todospg.Item{}

	co := coParentIds(ctx, accountId)
	query := `select t.*, a.name as owner_name from todos t join accounts a on a.id = t.account_id where t.account_id = ?`
	args := []interface{}{accountId}
	if len(co) > 0 {
		query += ` or (t.shared = true and t.account_id in ?)`
		args = append(args, co)
	}
	query += ` limit 800`
	own, err := gormdb.CustomSelect[todoRow](ctx, query, args...)
	if err != nil {
		return res, err
	}
	var done []todoRow
	for _, t := range own {
		if t.Done {
			done = append(done, t)
			continue
		}
		res.Items = append(res.Items, ownItem(t, accountId))
	}

	snoozed := activeSnoozes(ctx, accountId)
	out, err := dashboard.Outstanding(ctx, accountId)
	if err != nil {
		return res, err
	}
	for _, p := range out.TodoPayments {
		it := todospg.Item{
			Id: "payment:" + p.Id, Source: todospg.SourcePayment, Title: orDefault(p.Title, "Płatność"),
			Subtitle: "Do zapłaty", DueDate: p.DueDate, AmountGrosze: p.AmountGrosze, PaymentId: p.Id, Mine: true,
		}
		if !snoozed[it.Id] {
			res.Items = append(res.Items, it)
		}
	}
	for _, e := range out.TodoEvents {
		it := todospg.Item{
			Id: "event:" + e.Id + ":" + e.ChildId, Source: todospg.SourceEvent, Title: orDefault(e.Title, "Wydarzenie"),
			Subtitle: "Zapisz dziecko na wydarzenie", DueDate: e.RegistrationDeadline, EventId: e.Id, ChildId: e.ChildId, Mine: true,
		}
		if !snoozed[it.Id] {
			res.Items = append(res.Items, it)
		}
	}
	for _, c := range out.TodoConsents {
		kind, title := todospg.SourceConsent, "Zgoda: "+c.Content
		if c.Kind == "poll" {
			kind, title = todospg.SourcePoll, "Ankieta: "+c.Content
		}
		it := todospg.Item{
			Id: kind + ":" + c.MessageId, Source: kind, Title: title, Subtitle: "Czeka na Twoją odpowiedź",
			ChannelId: c.ChannelId, MessageId: c.MessageId, Mine: true,
		}
		if !snoozed[it.Id] {
			res.Items = append(res.Items, it)
		}
	}

	// Open items: dated ones by due date (important first on ties), then undated.
	sort.SliceStable(res.Items, func(i, j int) bool {
		a, b := res.Items[i], res.Items[j]
		switch {
		case a.DueDate == nil && b.DueDate == nil:
			return a.Important && !b.Important
		case a.DueDate == nil:
			return false
		case b.DueDate == nil:
			return true
		}
		if a.DueDate.Equal(*b.DueDate) {
			return a.Important && !b.Important
		}
		return a.DueDate.Before(*b.DueDate)
	})

	sort.Slice(done, func(i, j int) bool { return doneTime(done[i].Todo).After(doneTime(done[j].Todo)) })
	if len(done) > 50 {
		done = done[:50]
	}
	for _, t := range done {
		res.Items = append(res.Items, ownItem(t, accountId))
	}
	return res, nil
}

func doneTime(t todospg.Todo) time.Time {
	if t.DoneAt != nil {
		return *t.DoneAt
	}
	return t.UpdateTime
}

func shortName(full string) string {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return "Drugi rodzic"
	}
	return parts[0]
}

func ownItem(t todoRow, viewerId string) todospg.Item {
	it := todospg.Item{
		Id: t.Id, Source: todospg.SourceOwn, Title: t.Title, Subtitle: t.Note, DueDate: t.DueDate, Done: t.Done, DoneAt: t.DoneAt,
		ChildId: t.ChildId, Important: t.Important, RepeatRule: t.RepeatRule, RemindAt: t.RemindAt, Shared: t.Shared, Mine: t.AccountId == viewerId,
	}
	if !it.Mine {
		it.OwnerName = shortName(t.OwnerName)
	}
	if t.Checklist != "" {
		_ = json.Unmarshal([]byte(t.Checklist), &it.Checklist)
	}
	return it
}

func orDefault(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func validate(title, note string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleRequired
	}
	if len([]rune(title)) > 200 {
		return ErrTitleTooLong
	}
	if len([]rune(note)) > 1000 {
		return ErrNoteTooLong
	}
	return nil
}

func parseDue(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

var (
	ErrRepeat = errors.New("repeat_rule must be one of: daily, weekly, monthly, yearly")
	ErrChild  = errors.New("unknown child")
	ErrKey    = errors.New("unknown automatic to-do key")
)

func validRepeat(rule string) bool {
	switch rule {
	case "", "daily", "weekly", "monthly", "yearly":
		return true
	}
	return false
}

func encodeChecklist(list []todospg.ChecklistEntry) (string, error) {
	if len(list) == 0 {
		return "", nil
	}
	if len(list) > 50 {
		return "", errors.New("checklist may have at most 50 entries")
	}
	for i := range list {
		list[i].Text = strings.TrimSpace(list[i].Text)
		if list[i].Text == "" || len([]rune(list[i].Text)) > 200 {
			return "", errors.New("checklist entries must be 1-200 characters")
		}
	}
	b, err := json.Marshal(list)
	return string(b), err
}

func isParentOf(ctx context.Context, accountId, childId string) bool {
	if childId == "" {
		return true
	}
	type row struct{ N int }
	rows, err := gormdb.CustomSelect[row](ctx, "select count(*) as n from parent_children where parent_id = ? and child_id = ?", accountId, childId)
	return err == nil && len(rows) > 0 && rows[0].N > 0
}

// Create adds a task for accountId.
func Create(ctx context.Context, accountId string, req todospg.CreateRequest) (res todospg.Todo, err error) {
	ctx = logger.Enter(ctx, "todos.Create", accountId, req)
	defer logger.Exit(ctx, "todos.Create", &res, &err)

	if err = validate(req.Title, req.Note); err != nil {
		return res, err
	}
	if !validRepeat(req.RepeatRule) {
		return res, ErrRepeat
	}
	if !isParentOf(ctx, accountId, req.ChildId) {
		return res, ErrChild
	}
	due, err := parseDue(req.DueDate)
	if err != nil {
		return res, err
	}
	remind, err := parseDue(req.RemindAt)
	if err != nil {
		return res, err
	}
	checklist, err := encodeChecklist(req.Checklist)
	if err != nil {
		return res, err
	}
	now := time.Now()
	res = todospg.Todo{
		Id: uuid.NewString(), AccountId: accountId, Title: strings.TrimSpace(req.Title), Note: strings.TrimSpace(req.Note),
		DueDate: due, ChildId: req.ChildId, Important: req.Important, RepeatRule: req.RepeatRule, RemindAt: remind,
		Shared: req.Shared, Checklist: checklist, CreateTime: now, UpdateTime: now,
	}
	if err = gormdb.Create(ctx, &res); err != nil {
		return res, err
	}
	return res, nil
}

func byId(id string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: id}}}}}
}

// access loads a task the caller may edit: their own, or one the other parent shared.
func access(ctx context.Context, accountId, id string) (todospg.Todo, error) {
	rows, err := gormdb.Select[todospg.Todo](ctx, byId(id))
	if err != nil {
		return todospg.Todo{}, err
	}
	if len(rows) == 0 {
		return todospg.Todo{}, ErrNotFound
	}
	t := rows[0]
	if t.AccountId == accountId || (t.Shared && contains(coParentIds(ctx, accountId), t.AccountId)) {
		return t, nil
	}
	return todospg.Todo{}, ErrNotFound
}

// nextOccurrence advances due by the repeat rule until it lies in the future.
func nextOccurrence(due time.Time, rule string, now time.Time) time.Time {
	next := due
	for i := 0; i < 400; i++ {
		switch rule {
		case "daily":
			next = next.AddDate(0, 0, 1)
		case "weekly":
			next = next.AddDate(0, 0, 7)
		case "monthly":
			next = next.AddDate(0, 1, 0)
		case "yearly":
			next = next.AddDate(1, 0, 0)
		default:
			return next
		}
		if next.After(now) {
			return next
		}
	}
	return next
}

// Update applies a partial edit; the owner and (for shared tasks) the other parent may edit. Ticking Done
// stamps DoneAt and, for a repeating task with a due date, creates the next occurrence.
func Update(ctx context.Context, accountId, id string, req todospg.UpdateRequest) (res todospg.Todo, err error) {
	ctx = logger.Enter(ctx, "todos.Update", accountId, id, req)
	defer logger.Exit(ctx, "todos.Update", &res, &err)

	t, err := access(ctx, accountId, id)
	if err != nil {
		return res, err
	}
	owner := t.AccountId == accountId
	if req.Title != nil {
		t.Title = strings.TrimSpace(*req.Title)
	}
	if req.Note != nil {
		t.Note = strings.TrimSpace(*req.Note)
	}
	if err = validate(t.Title, t.Note); err != nil {
		return res, err
	}
	if req.DueDate != nil {
		if t.DueDate, err = parseDue(*req.DueDate); err != nil {
			return res, err
		}
	}
	if req.RemindAt != nil {
		if t.RemindAt, err = parseDue(*req.RemindAt); err != nil {
			return res, err
		}
		t.Reminded = false
	}
	if req.Important != nil {
		t.Important = *req.Important
	}
	if req.RepeatRule != nil {
		if !validRepeat(*req.RepeatRule) {
			return res, ErrRepeat
		}
		t.RepeatRule = *req.RepeatRule
	}
	if req.ChildId != nil {
		if !isParentOf(ctx, t.AccountId, *req.ChildId) {
			return res, ErrChild
		}
		t.ChildId = *req.ChildId
	}
	if req.Shared != nil && owner {
		t.Shared = *req.Shared
	}
	if req.Checklist != nil {
		if t.Checklist, err = encodeChecklist(*req.Checklist); err != nil {
			return res, err
		}
	}
	now := time.Now()
	completedNow := false
	if req.Done != nil && *req.Done != t.Done {
		t.Done = *req.Done
		if t.Done {
			t.DoneAt = &now
			completedNow = true
		} else {
			t.DoneAt = nil
		}
	}
	t.UpdateTime = now
	if err = gormdb.Update(ctx, &t, byId(id)); err != nil {
		return res, err
	}
	if completedNow && t.RepeatRule != "" && t.DueDate != nil {
		next := nextOccurrence(*t.DueDate, t.RepeatRule, now)
		clone := t
		clone.Id, clone.Done, clone.DoneAt, clone.Reminded = uuid.NewString(), false, nil, false
		clone.DueDate, clone.CreateTime, clone.UpdateTime = &next, now, now
		if t.RemindAt != nil {
			r := t.RemindAt.Add(next.Sub(*t.DueDate))
			clone.RemindAt = &r
		}
		if t.Checklist != "" { // a fresh checklist: everything unticked
			var entries []todospg.ChecklistEntry
			if json.Unmarshal([]byte(t.Checklist), &entries) == nil {
				for i := range entries {
					entries[i].Done = false
				}
				clone.Checklist, _ = encodeChecklist(entries)
			}
		}
		if err = gormdb.Create(ctx, &clone); err != nil {
			return t, err
		}
	}
	return t, nil
}

// Delete removes an own task (owner only).
func Delete(ctx context.Context, accountId, id string) (err error) {
	ctx = logger.Enter(ctx, "todos.Delete", accountId, id)
	defer logger.Exit(ctx, "todos.Delete", &err)

	existing, err := gormdb.Select[todospg.Todo](ctx, byIdAndOwner(id, accountId))
	if err != nil {
		return err
	}
	if len(existing) == 0 {
		return ErrNotFound
	}
	_, err = gormdb.Delete[todospg.Todo](ctx, byIdAndOwner(id, accountId))
	return err
}

// ---- snoozing automatic to-dos ----

func activeSnoozes(ctx context.Context, accountId string) map[string]bool {
	out := map[string]bool{}
	rows, err := gormdb.CustomSelect[todospg.Snooze](ctx, "select * from todo_snoozes where account_id = ? and until > ?", accountId, time.Now())
	if err != nil {
		return out
	}
	for _, r := range rows {
		out[r.Key] = true
	}
	return out
}

// Snooze hides an automatic to-do until the given time ("Odłóż" / "Ukryj").
func Snooze(ctx context.Context, accountId string, req todospg.SnoozeRequest) error {
	ok := false
	for _, prefix := range []string{"payment:", "event:", "poll:", "consent:"} {
		if strings.HasPrefix(req.Key, prefix) {
			ok = true
		}
	}
	if !ok || len(req.Key) > 200 {
		return ErrKey
	}
	until, err := time.Parse(time.RFC3339, req.Until)
	if err != nil {
		return err
	}
	return gormdb.GormDB.WithContext(ctx).Exec(`insert into todo_snoozes (account_id, key, until) values (?, ?, ?)
 on conflict (account_id, key) do update set until = excluded.until`, accountId, req.Key, until).Error
}

// Unsnooze brings a hidden automatic to-do back.
func Unsnooze(ctx context.Context, accountId, key string) error {
	return gormdb.GormDB.WithContext(ctx).Exec("delete from todo_snoozes where account_id = ? and key = ?", accountId, key).Error
}

// ---- notification preferences (opt-in) ----

func GetPrefs(ctx context.Context, accountId string) todospg.Prefs {
	rows, err := gormdb.CustomSelect[todospg.Prefs](ctx, "select * from todo_prefs where account_id = ?", accountId)
	if err != nil || len(rows) == 0 {
		return todospg.Prefs{AccountId: accountId}
	}
	return rows[0]
}

func SetPrefs(ctx context.Context, accountId string, digest, dueReminders bool) (todospg.Prefs, error) {
	err := gormdb.GormDB.WithContext(ctx).Exec(`insert into todo_prefs (account_id, digest, due_reminders) values (?, ?, ?)
 on conflict (account_id) do update set digest = excluded.digest, due_reminders = excluded.due_reminders`, accountId, digest, dueReminders).Error
	return todospg.Prefs{AccountId: accountId, Digest: digest, DueReminders: dueReminders}, err
}
