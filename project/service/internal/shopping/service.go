package shopping

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	sp "github.com/Johncode248/Rodzic-Chat/project/common/shopping"
	"github.com/google/uuid"
)

var (
	ErrInvalid  = errors.New("invalid data")
	ErrNotFound = errors.New("not found")
)

const maxItemsPerList = 300

// family = the account plus its co-parents (parents sharing at least one child).
func family(ctx context.Context, accountId string) []string {
	type row struct{ ParentId string }
	rows, _ := gormdb.CustomSelect[row](ctx, `select distinct pc2.parent_id from parent_children pc1
 join parent_children pc2 on pc1.child_id = pc2.child_id and pc2.parent_id <> pc1.parent_id where pc1.parent_id = ?`, accountId)
	ids := []string{accountId}
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

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

func shortName(ctx context.Context, id string) string {
	type row struct{ Name string }
	rows, _ := gormdb.CustomSelect[row](ctx, "select name from accounts where id = ?", id)
	if len(rows) == 0 {
		return ""
	}
	if f := strings.Fields(rows[0].Name); len(f) > 0 {
		return f[0]
	}
	return ""
}

func defaultName() string { return "Zakupy " + time.Now().Format("2.01") }

// loadList returns a list the account may access (its own or a co-parent's).
func loadList(ctx context.Context, accountId, id string) (sp.List, []string, error) {
	rows, err := gormdb.CustomSelect[sp.List](ctx, "select * from shopping_lists where id = ?", id)
	fam := family(ctx, accountId)
	if err != nil || len(rows) == 0 || !contains(fam, rows[0].AccountId) {
		return sp.List{}, nil, ErrNotFound
	}
	return rows[0], fam, nil
}

func summaries(ctx context.Context, lists []sp.List) []sp.ListSummary {
	out := make([]sp.ListSummary, 0, len(lists))
	if len(lists) == 0 {
		return out
	}
	ids := make([]string, 0, len(lists))
	for _, l := range lists {
		ids = append(ids, l.Id)
	}
	type cnt struct {
		ListId string
		Total  int
		Done   int
	}
	rows, _ := gormdb.CustomSelect[cnt](ctx, "select list_id, count(*) as total, count(*) filter (where done) as done from shopping_items where list_id in ? group by list_id", ids)
	by := map[string]cnt{}
	for _, r := range rows {
		by[r.ListId] = r
	}
	for _, l := range lists {
		c := by[l.Id]
		out = append(out, sp.ListSummary{List: l, Total: c.Total, Done: c.Done})
	}
	return out
}

// ---- lists ----

// Lists returns the family's lists with the given status ("active" newest first, "completed" by finish date).
func Lists(ctx context.Context, accountId, status string) ([]sp.ListSummary, error) {
	if status != sp.StatusActive && status != sp.StatusCompleted {
		return nil, ErrInvalid
	}
	order := "create_time desc"
	limit := 100
	if status == sp.StatusCompleted {
		order = "completed_time desc nulls last"
	}
	rows, err := gormdb.CustomSelect[sp.List](ctx, fmt.Sprintf("select * from shopping_lists where account_id in ? and status = ? order by %s limit %d", order, limit), family(ctx, accountId), status)
	if err != nil {
		return nil, err
	}
	return summaries(ctx, rows), nil
}

func CreateList(ctx context.Context, accountId string, req sp.CreateListRequest) (sp.ListSummary, error) {
	name := clip(req.Name, 60)
	if name == "" {
		name = defaultName()
	}
	l := sp.List{Id: uuid.NewString(), AccountId: accountId, Name: name, Status: sp.StatusActive, CreatedByName: shortName(ctx, accountId), CreateTime: time.Now()}
	if err := gormdb.Create(ctx, &l); err != nil {
		return sp.ListSummary{}, err
	}
	return sp.ListSummary{List: l}, nil
}

func UpdateList(ctx context.Context, accountId, id string, req sp.UpdateListRequest) (sp.ListSummary, error) {
	l, _, err := loadList(ctx, accountId, id)
	if err != nil {
		return sp.ListSummary{}, err
	}
	if req.Name != nil {
		if n := clip(*req.Name, 60); n != "" {
			l.Name = n
		}
	}
	if req.Status != nil && *req.Status != l.Status {
		switch *req.Status {
		case sp.StatusCompleted:
			now := time.Now()
			l.Status, l.CompletedTime = sp.StatusCompleted, &now
		case sp.StatusActive:
			l.Status, l.CompletedTime = sp.StatusActive, nil
		default:
			return sp.ListSummary{}, ErrInvalid
		}
	}
	if err := gormdb.GormDB.WithContext(ctx).Exec("update shopping_lists set name = ?, status = ?, completed_time = ? where id = ?", l.Name, l.Status, l.CompletedTime, id).Error; err != nil {
		return sp.ListSummary{}, err
	}
	return summaries(ctx, []sp.List{l})[0], nil
}

func DeleteList(ctx context.Context, accountId, id string) error {
	if _, _, err := loadList(ctx, accountId, id); err != nil {
		return err
	}
	db := gormdb.GormDB.WithContext(ctx)
	db.Exec("delete from shopping_items where list_id = ?", id)
	return db.Exec("delete from shopping_lists where id = ?", id).Error
}

// CopyList starts a new active list from an old one: every product again, nothing ticked.
func CopyList(ctx context.Context, accountId, id string, req sp.CopyListRequest) (sp.ListSummary, error) {
	src, _, err := loadList(ctx, accountId, id)
	if err != nil {
		return sp.ListSummary{}, err
	}
	name := clip(req.Name, 60)
	if name == "" {
		name = defaultName()
	}
	nl, err := CreateList(ctx, accountId, sp.CreateListRequest{Name: name})
	if err != nil {
		return nl, err
	}
	items, _ := gormdb.CustomSelect[sp.Item](ctx, "select * from shopping_items where list_id = ? order by create_time limit ?", src.Id, maxItemsPerList)
	who := shortName(ctx, accountId)
	now := time.Now()
	for i, it := range items {
		c := sp.Item{Id: uuid.NewString(), ListId: nl.Id, AccountId: accountId, Name: it.Name, Quantity: it.Quantity, AddedByName: who, CreateTime: now.Add(time.Duration(i) * time.Millisecond)}
		if err := gormdb.Create(ctx, &c); err != nil {
			return nl, err
		}
	}
	nl.Total = len(items)
	return nl, nil
}

// ---- one list ----

func Detail(ctx context.Context, accountId, id string) (sp.ListDetail, error) {
	l, fam, err := loadList(ctx, accountId, id)
	if err != nil {
		return sp.ListDetail{}, err
	}
	items, err := gormdb.CustomSelect[sp.Item](ctx, "select * from shopping_items where list_id = ? order by done, create_time desc limit ?", id, maxItemsPerList)
	if err != nil {
		return sp.ListDetail{}, err
	}
	if items == nil {
		items = []sp.Item{}
	}
	out := sp.ListDetail{List: summaries(ctx, []sp.List{l})[0], Items: items, Suggestions: []string{}}
	if l.Status == sp.StatusActive {
		on := map[string]bool{}
		for _, it := range items {
			on[strings.ToLower(it.Name)] = true
		}
		for _, h := range history(ctx, fam, "", 40) {
			if !on[strings.ToLower(h.Name)] && len(out.Suggestions) < 12 {
				out.Suggestions = append(out.Suggestions, h.Name)
			}
		}
	}
	return out, nil
}

// ---- items ----

func AddItems(ctx context.Context, accountId string, req sp.AddItemsRequest) ([]sp.Item, error) {
	l, _, err := loadList(ctx, accountId, req.ListId)
	if err != nil {
		return nil, err
	}
	if len(req.Items) == 0 || len(req.Items) > 100 {
		return nil, ErrInvalid
	}
	type cnt struct{ N int }
	if c, _ := gormdb.CustomSelect[cnt](ctx, "select count(*) as n from shopping_items where list_id = ?", l.Id); len(c) > 0 && c[0].N+len(req.Items) > maxItemsPerList {
		return nil, fmt.Errorf("%w: lista może mieć maksymalnie %d pozycji", ErrInvalid, maxItemsPerList)
	}
	who := shortName(ctx, accountId)
	now := time.Now()
	out := make([]sp.Item, 0, len(req.Items))
	for i, in := range req.Items {
		name := clip(in.Name, 80)
		if name == "" {
			continue
		}
		it := sp.Item{Id: uuid.NewString(), ListId: l.Id, AccountId: accountId, Name: name, Quantity: clip(in.Quantity, 30), AddedByName: who, CreateTime: now.Add(time.Duration(i) * time.Millisecond)}
		if err := gormdb.Create(ctx, &it); err != nil {
			return out, err
		}
		out = append(out, it)
	}
	if len(out) == 0 {
		return nil, ErrInvalid
	}
	return out, nil
}

func loadItem(ctx context.Context, accountId, id string) (sp.Item, error) {
	rows, err := gormdb.CustomSelect[sp.Item](ctx, "select * from shopping_items where id = ?", id)
	if err != nil || len(rows) == 0 {
		return sp.Item{}, ErrNotFound
	}
	if _, _, err := loadList(ctx, accountId, rows[0].ListId); err != nil {
		return sp.Item{}, ErrNotFound
	}
	return rows[0], nil
}

func UpdateItem(ctx context.Context, accountId, id string, req sp.UpdateItemRequest) (sp.Item, error) {
	it, err := loadItem(ctx, accountId, id)
	if err != nil {
		return it, err
	}
	if req.Name != nil {
		if n := clip(*req.Name, 80); n != "" {
			it.Name = n
		}
	}
	if req.Quantity != nil {
		it.Quantity = clip(*req.Quantity, 30)
	}
	if req.Done != nil && *req.Done != it.Done {
		it.Done = *req.Done
		if it.Done {
			now := time.Now()
			it.DoneAt, it.DoneBy = &now, shortName(ctx, accountId)
		} else {
			it.DoneAt, it.DoneBy = nil, ""
		}
	}
	return it, gormdb.GormDB.WithContext(ctx).Exec("update shopping_items set name = ?, quantity = ?, done = ?, done_by = ?, done_at = ? where id = ?",
		it.Name, it.Quantity, it.Done, it.DoneBy, it.DoneAt, id).Error
}

func DeleteItem(ctx context.Context, accountId, id string) error {
	if _, err := loadItem(ctx, accountId, id); err != nil {
		return err
	}
	return gormdb.GormDB.WithContext(ctx).Exec("delete from shopping_items where id = ?", id).Error
}

// ---- history ----

// history: distinct products the family ever had on a list, most frequent first (case-insensitive), optional name filter.
func history(ctx context.Context, fam []string, q string, limit int) []sp.HistoryEntry {
	type row struct {
		Name         string
		Count        int
		LastQuantity string
	}
	like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
	rows, _ := gormdb.CustomSelect[row](ctx, `select min(name) as name, count(*) as count,
 (array_agg(quantity order by create_time desc))[1] as last_quantity
 from shopping_items where account_id in ? and lower(name) like ? group by lower(name)
 order by count(*) desc, max(create_time) desc limit ?`, fam, like, limit)
	out := make([]sp.HistoryEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, sp.HistoryEntry{Name: r.Name, Count: r.Count, LastQuantity: r.LastQuantity})
	}
	return out
}

func History(ctx context.Context, accountId, q string) []sp.HistoryEntry {
	return history(ctx, family(ctx, accountId), q, 80)
}
