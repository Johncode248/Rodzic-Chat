package todos

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	todospg "github.com/Johncode248/Rodzic-Chat/project/common/todos"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

var warsaw = func() *time.Location {
	if loc, err := time.LoadLocation("Europe/Warsaw"); err == nil {
		return loc
	}
	return time.UTC
}()

// RunScheduler sends the push notifications of the "Do zrobienia" feature once a minute:
//   - a task's own reminder (RemindAt) -- to its owner (and the other parent when the task is shared);
//   - the opt-in 8:00 digest ("Dziś masz N zadań") and the opt-in 18:00 "jutro" reminder.
//
// Every send is guarded by an atomic UPDATE ... WHERE (reminded = false / last day <> today), so several
// backend instances never notify twice. Nothing is sent to anyone who has not opted in (digests) or set a reminder.
func RunScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("todos scheduler: recovered: %v", r)
					}
				}()
				sendReminders(ctx)
				sendDigests(ctx)
			}()
		}
	}
}

func sendReminders(ctx context.Context) {
	rows, err := gormdb.CustomSelect[todospg.Todo](ctx, "select * from todos where remind_at is not null and remind_at <= ? and reminded = false and done = false limit 200", time.Now())
	if err != nil {
		return
	}
	for _, t := range rows {
		res := gormdb.GormDB.WithContext(ctx).Exec("update todos set reminded = true where id = ? and reminded = false", t.Id)
		if res.Error != nil || res.RowsAffected != 1 {
			continue
		}
		targets := []string{t.AccountId}
		if t.Shared {
			targets = append(targets, coParentIds(ctx, t.AccountId)...)
		}
		push.SendToAccounts(ctx, targets, "Przypomnienie", t.Title, map[string]string{"type": "todo_reminder", "todo_id": t.Id})
	}
}

func sendDigests(ctx context.Context) {
	now := time.Now().In(warsaw)
	today := now.Format("2006-01-02")

	if now.Hour() == 8 {
		rows, err := gormdb.CustomSelect[todospg.Prefs](ctx, "select * from todo_prefs where digest = true and last_digest_day <> ? or (digest = true and last_digest_day is null)", today)
		if err == nil {
			for _, p := range rows {
				res := gormdb.GormDB.WithContext(ctx).Exec("update todo_prefs set last_digest_day = ? where account_id = ? and (last_digest_day is null or last_digest_day <> ?)", today, p.AccountId, today)
				if res.Error != nil || res.RowsAffected != 1 {
					continue
				}
				endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, warsaw)
				n := openDueCount(ctx, p.AccountId, endOfToday)
				if n > 0 {
					push.SendToAccounts(ctx, []string{p.AccountId}, "Dziś do zrobienia", fmt.Sprintf("Masz %d %s na dziś.", n, taskWord(n)), map[string]string{"type": "todo_digest"})
				}
			}
		}
	}

	if now.Hour() == 18 {
		rows, err := gormdb.CustomSelect[todospg.Prefs](ctx, "select * from todo_prefs where due_reminders = true and (last_reminder_day is null or last_reminder_day <> ?)", today)
		if err == nil {
			for _, p := range rows {
				res := gormdb.GormDB.WithContext(ctx).Exec("update todo_prefs set last_reminder_day = ? where account_id = ? and (last_reminder_day is null or last_reminder_day <> ?)", today, p.AccountId, today)
				if res.Error != nil || res.RowsAffected != 1 {
					continue
				}
				tomorrowEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 23, 59, 59, 0, warsaw)
				todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, warsaw)
				n := openDueCount(ctx, p.AccountId, tomorrowEnd) - openDueCount(ctx, p.AccountId, todayEnd)
				if n > 0 {
					push.SendToAccounts(ctx, []string{p.AccountId}, "Jutro do zrobienia", fmt.Sprintf("Na jutro masz %d %s.", n, taskWord(n)), map[string]string{"type": "todo_tomorrow"})
				}
			}
		}
	}
}

// openDueCount counts the account's open tasks (own + shared by the other parent) due up to the given time.
func openDueCount(ctx context.Context, accountId string, until time.Time) int {
	type row struct{ N int }
	co := coParentIds(ctx, accountId)
	query := "select count(*) as n from todos where done = false and due_date is not null and due_date <= ? and (account_id = ?"
	args := []interface{}{until, accountId}
	if len(co) > 0 {
		query += " or (shared = true and account_id in ?)"
		args = append(args, co)
	}
	query += ")"
	rows, err := gormdb.CustomSelect[row](ctx, query, args...)
	if err != nil || len(rows) == 0 {
		return 0
	}
	return rows[0].N
}

func taskWord(n int) string {
	switch {
	case n == 1:
		return "zadanie"
	case n%10 >= 2 && n%10 <= 4 && !(n%100 >= 12 && n%100 <= 14):
		return "zadania"
	}
	return "zadań"
}
