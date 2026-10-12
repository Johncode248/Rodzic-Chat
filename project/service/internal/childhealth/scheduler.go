package childhealth

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	hp "github.com/Johncode248/Rodzic-Chat/project/common/childhealth"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

// RunScheduler pushes medication reminders to every parent of the child once a minute: at the dose time, and once
// more 15 minutes later if nobody has marked it "podano". Each step is claimed with an atomic UPDATE so several
// backend instances never notify twice.
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
						log.Printf("childhealth scheduler: recovered: %v", r)
					}
				}()
				remind(ctx, 0, 0, 1, "Czas na lek")
				remind(ctx, 15*time.Minute, 1, 2, "Przypomnienie: lek nadal niepodany")
			}()
		}
	}
}

// remind notifies doses that are pending, due at least [after] ago and still at remind_count = from.
func remind(ctx context.Context, after time.Duration, from, to int, title string) {
	rows, err := gormdb.CustomSelect[hp.Dose](ctx, "select * from child_doses where status = 'pending' and remind_count = ? and due_time <= ? limit 100", from, time.Now().Add(-after))
	if err != nil {
		return
	}
	for _, d := range rows {
		res := gormdb.GormDB.WithContext(ctx).Exec("update child_doses set remind_count = ? where id = ? and remind_count = ? and status = 'pending'", to, d.Id, from)
		if res.Error != nil || res.RowsAffected != 1 {
			continue
		}
		ms, _ := gormdb.CustomSelect[hp.Medication](ctx, "select * from child_medications where id = ? and active = true", d.MedicationId)
		if len(ms) == 0 {
			continue
		}
		body := childName(ctx, d.ChildId) + ": " + strings.TrimSpace(ms[0].Name+" "+ms[0].Dose)
		push.SendToAccounts(ctx, parentIds(ctx, d.ChildId), title, body, map[string]string{"type": "dose_due", "child_id": d.ChildId, "dose_id": d.Id})
	}
}
