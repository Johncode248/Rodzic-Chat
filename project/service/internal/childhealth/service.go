package childhealth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	hp "github.com/Johncode248/Rodzic-Chat/project/common/childhealth"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
	"github.com/google/uuid"
)

var (
	ErrForbidden   = errors.New("not allowed")
	ErrNotFound    = errors.New("not found")
	ErrInvalid     = errors.New("invalid data")
	ErrEpisodeOpen = errors.New("this child already has an open illness -- end it first")
)

// ---- access: only parents linked to the child ----

func IsParent(ctx context.Context, accountId, childId string) bool {
	if accountId == "" || childId == "" {
		return false
	}
	type row struct{ N int }
	rows, err := gormdb.CustomSelect[row](ctx, "select count(*) as n from parent_children where parent_id = ? and child_id = ?", accountId, childId)
	return err == nil && len(rows) > 0 && rows[0].N > 0
}

func parentIds(ctx context.Context, childId string) []string {
	type row struct{ ParentId string }
	rows, _ := gormdb.CustomSelect[row](ctx, "select distinct parent_id from parent_children where child_id = ?", childId)
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.ParentId)
	}
	return out
}

func others(ids []string, me string) []string {
	out := []string{}
	for _, id := range ids {
		if id != me {
			out = append(out, id)
		}
	}
	return out
}

func accountName(ctx context.Context, id string) string {
	type row struct{ Name string }
	rows, _ := gormdb.CustomSelect[row](ctx, "select name from accounts where id = ?", id)
	if len(rows) == 0 {
		return "Rodzic"
	}
	parts := strings.Fields(rows[0].Name)
	if len(parts) == 0 {
		return "Rodzic"
	}
	return parts[0]
}

func childName(ctx context.Context, id string) string {
	type row struct{ FirstName string }
	rows, _ := gormdb.CustomSelect[row](ctx, "select first_name from children where id = ?", id)
	if len(rows) == 0 {
		return "Dziecko"
	}
	return rows[0].FirstName
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

func parseTime(raw string, fallback time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}

// ---- health card ----

var bloodTypes = map[string]bool{"": true, "0+": true, "0-": true, "A+": true, "A-": true, "B+": true, "B-": true, "AB+": true, "AB-": true}

func GetOverview(ctx context.Context, accountId, childId string) (hp.Overview, error) {
	if !IsParent(ctx, accountId, childId) {
		return hp.Overview{}, ErrForbidden
	}
	out := hp.Overview{Profile: hp.Profile{ChildId: childId}, Allergies: []hp.Allergy{}, Doctors: []hp.Doctor{}, Measurements: []hp.Measurement{}}
	type card struct{ BloodType, HealthNote string }
	if rows, err := gormdb.CustomSelect[card](ctx, "select coalesce(blood_type,'') as blood_type, coalesce(health_note,'') as health_note from children where id = ?", childId); err == nil && len(rows) > 0 {
		out.Profile.BloodType, out.Profile.Note = rows[0].BloodType, rows[0].HealthNote
	}
	if rows, err := gormdb.CustomSelect[hp.Allergy](ctx, "select * from child_allergies where child_id = ? order by create_time", childId); err == nil && rows != nil {
		out.Allergies = rows
	}
	if rows, err := gormdb.CustomSelect[hp.Doctor](ctx, "select * from child_doctors where child_id = ? order by create_time", childId); err == nil && rows != nil {
		out.Doctors = rows
	}
	if rows, err := gormdb.CustomSelect[hp.Measurement](ctx, "select * from child_measurements where child_id = ? order by date desc limit 60", childId); err == nil && rows != nil {
		out.Measurements = rows
	}
	if eps, err := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where child_id = ? and end_time is null order by start_time desc limit 1", childId); err == nil && len(eps) > 0 {
		out.ActiveTitle = eps[0].Title
	}
	type nd struct {
		DoseId  string
		Name    string
		Dose    string
		DueTime time.Time
	}
	if rows, err := gormdb.CustomSelect[nd](ctx, `select d.id as dose_id, m.name, m.dose, d.due_time from child_doses d join child_medications m on m.id = d.medication_id
 where d.child_id = ? and d.status = 'pending' and m.active = true order by d.due_time limit 1`, childId); err == nil && len(rows) > 0 {
		out.NextDose = &hp.NextDose{DoseId: rows[0].DoseId, Name: rows[0].Name, Dose: rows[0].Dose, DueTime: rows[0].DueTime, Overdue: rows[0].DueTime.Before(time.Now())}
	}
	return out, nil
}

func SaveProfile(ctx context.Context, accountId string, req hp.ProfileRequest) error {
	if !IsParent(ctx, accountId, req.ChildId) {
		return ErrForbidden
	}
	if !bloodTypes[req.BloodType] {
		return ErrInvalid
	}
	return gormdb.GormDB.WithContext(ctx).Exec("update children set blood_type = ?, health_note = ?, update_time = ? where id = ?",
		req.BloodType, clip(req.Note, 1000), time.Now(), req.ChildId).Error
}

func AddAllergy(ctx context.Context, accountId string, req hp.AllergyRequest) (hp.Allergy, error) {
	if !IsParent(ctx, accountId, req.ChildId) {
		return hp.Allergy{}, ErrForbidden
	}
	name := clip(req.Name, 80)
	sev := req.Severity
	if name == "" || (sev != "mild" && sev != "moderate" && sev != "severe") {
		return hp.Allergy{}, ErrInvalid
	}
	a := hp.Allergy{Id: uuid.NewString(), ChildId: req.ChildId, Name: name, Severity: sev, Reaction: clip(req.Reaction, 200), CreateTime: time.Now()}
	return a, gormdb.Create(ctx, &a)
}

func AddDoctor(ctx context.Context, accountId string, req hp.DoctorRequest) (hp.Doctor, error) {
	if !IsParent(ctx, accountId, req.ChildId) {
		return hp.Doctor{}, ErrForbidden
	}
	name := clip(req.Name, 100)
	if name == "" {
		return hp.Doctor{}, ErrInvalid
	}
	d := hp.Doctor{Id: uuid.NewString(), ChildId: req.ChildId, Name: name, Specialty: clip(req.Specialty, 80), Phone: clip(req.Phone, 30), Address: clip(req.Address, 200), CreateTime: time.Now()}
	return d, gormdb.Create(ctx, &d)
}

func AddMeasurement(ctx context.Context, accountId string, req hp.MeasurementRequest) (hp.Measurement, error) {
	if !IsParent(ctx, accountId, req.ChildId) {
		return hp.Measurement{}, ErrForbidden
	}
	date, err := parseTime(req.Date, time.Now())
	if err != nil || (req.WeightKg <= 0 && req.HeightCm <= 0) || req.WeightKg > 200 || req.HeightCm > 230 || req.WeightKg < 0 || req.HeightCm < 0 {
		return hp.Measurement{}, ErrInvalid
	}
	m := hp.Measurement{Id: uuid.NewString(), ChildId: req.ChildId, Date: date, WeightKg: req.WeightKg, HeightCm: req.HeightCm, CreateTime: time.Now()}
	return m, gormdb.Create(ctx, &m)
}

// Delete removes a health-card row (allergies, doctors, measurements) the caller may access.
func Delete(ctx context.Context, accountId, table, id string) error {
	childId := ""
	switch table {
	case "child_allergies", "child_doctors", "child_measurements":
		type row struct{ ChildId string }
		rows, err := gormdb.CustomSelect[row](ctx, "select child_id from "+table+" where id = ?", id)
		if err != nil || len(rows) == 0 {
			return ErrNotFound
		}
		childId = rows[0].ChildId
	default:
		return ErrInvalid
	}
	if !IsParent(ctx, accountId, childId) {
		return ErrForbidden
	}
	return gormdb.GormDB.WithContext(ctx).Exec("delete from "+table+" where id = ?", id).Error
}

// ---- illness diary ----

func GetIllness(ctx context.Context, accountId, childId string) (hp.IllnessOverview, error) {
	if !IsParent(ctx, accountId, childId) {
		return hp.IllnessOverview{}, ErrForbidden
	}
	out := hp.IllnessOverview{Past: []hp.Episode{}}
	if eps, err := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where child_id = ? and end_time is null order by start_time desc limit 1", childId); err == nil && len(eps) > 0 {
		cur := &hp.Current{Episode: eps[0], Entries: []hp.Entry{}, Medications: []hp.MedicationView{}}
		if rows, err := gormdb.CustomSelect[hp.Entry](ctx, "select * from child_episode_entries where episode_id = ? order by entry_time desc limit 300", cur.Episode.Id); err == nil && rows != nil {
			cur.Entries = rows
		}
		meds, _ := gormdb.CustomSelect[hp.Medication](ctx, "select * from child_medications where child_id = ? and active = true and (episode_id = ? or episode_id = '') order by create_time", childId, cur.Episode.Id)
		for _, m := range meds {
			doses, _ := gormdb.CustomSelect[hp.Dose](ctx, "select * from child_doses where medication_id = ? order by due_time limit 200", m.Id)
			mv := hp.MedicationView{Medication: m, Doses: doses}
			if mv.Doses == nil {
				mv.Doses = []hp.Dose{}
			}
			for _, d := range doses {
				if d.Status == hp.DoseGiven && d.GivenTime != nil && (mv.LastGiven == nil || d.GivenTime.After(*mv.LastGiven)) {
					t := *d.GivenTime
					mv.LastGiven = &t
				}
			}
			cur.Medications = append(cur.Medications, mv)
		}
		out.Current = cur
	}
	if rows, err := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where child_id = ? and end_time is not null order by start_time desc limit 15", childId); err == nil && rows != nil {
		out.Past = rows
	}
	return out, nil
}

func GetEpisode(ctx context.Context, accountId, episodeId string) (hp.EpisodeDetail, error) {
	eps, err := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where id = ?", episodeId)
	if err != nil || len(eps) == 0 {
		return hp.EpisodeDetail{}, ErrNotFound
	}
	if !IsParent(ctx, accountId, eps[0].ChildId) {
		return hp.EpisodeDetail{}, ErrForbidden
	}
	entries, _ := gormdb.CustomSelect[hp.Entry](ctx, "select * from child_episode_entries where episode_id = ? order by entry_time", episodeId)
	if entries == nil {
		entries = []hp.Entry{}
	}
	return hp.EpisodeDetail{Episode: eps[0], Entries: entries}, nil
}

func StartEpisode(ctx context.Context, accountId string, req hp.EpisodeRequest) (hp.Episode, error) {
	if !IsParent(ctx, accountId, req.ChildId) {
		return hp.Episode{}, ErrForbidden
	}
	title := clip(req.Title, 100)
	if title == "" {
		return hp.Episode{}, ErrInvalid
	}
	if open, _ := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where child_id = ? and end_time is null", req.ChildId); len(open) > 0 {
		return hp.Episode{}, ErrEpisodeOpen
	}
	e := hp.Episode{Id: uuid.NewString(), ChildId: req.ChildId, Title: title, StartTime: time.Now(), CreatedBy: accountId, CreateTime: time.Now()}
	if err := gormdb.Create(ctx, &e); err != nil {
		return e, err
	}
	notifyOthers(ctx, req.ChildId, accountId, childName(ctx, req.ChildId)+" jest chore", accountName(ctx, accountId)+" zapisał(a): "+title, "illness_start")
	return e, nil
}

func EndEpisode(ctx context.Context, accountId, episodeId string) error {
	eps, err := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where id = ?", episodeId)
	if err != nil || len(eps) == 0 {
		return ErrNotFound
	}
	if !IsParent(ctx, accountId, eps[0].ChildId) {
		return ErrForbidden
	}
	now := time.Now()
	db := gormdb.GormDB.WithContext(ctx)
	if err := db.Exec("update child_episodes set end_time = ? where id = ? and end_time is null", now, episodeId).Error; err != nil {
		return err
	}
	// Stop the schedules of this illness: nobody should be reminded about medicine for a finished illness.
	db.Exec("update child_medications set active = false where episode_id = ?", episodeId)
	db.Exec("delete from child_doses where status = 'pending' and medication_id in (select id from child_medications where episode_id = ?)", episodeId)
	return nil
}

func AddEntry(ctx context.Context, accountId string, req hp.EntryRequest) (hp.Entry, error) {
	eps, err := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where id = ?", req.EpisodeId)
	if err != nil || len(eps) == 0 {
		return hp.Entry{}, ErrNotFound
	}
	if !IsParent(ctx, accountId, eps[0].ChildId) {
		return hp.Entry{}, ErrForbidden
	}
	kind := req.Kind
	if kind != hp.KindTemperature && kind != hp.KindSymptom && kind != hp.KindNote {
		return hp.Entry{}, ErrInvalid
	}
	if kind == hp.KindTemperature && (req.Value < 34 || req.Value > 43) {
		return hp.Entry{}, ErrInvalid
	}
	if kind != hp.KindTemperature && clip(req.Text, 300) == "" {
		return hp.Entry{}, ErrInvalid
	}
	at, err := parseTime(req.EntryTime, time.Now())
	if err != nil {
		return hp.Entry{}, ErrInvalid
	}
	e := hp.Entry{Id: uuid.NewString(), EpisodeId: req.EpisodeId, ChildId: eps[0].ChildId, Kind: kind, Text: clip(req.Text, 300), Value: req.Value, EntryTime: at, CreatedBy: accountId, CreatedByName: accountName(ctx, accountId)}
	return e, gormdb.Create(ctx, &e)
}

func DeleteEntry(ctx context.Context, accountId, id string) error {
	rows, err := gormdb.CustomSelect[hp.Entry](ctx, "select * from child_episode_entries where id = ?", id)
	if err != nil || len(rows) == 0 {
		return ErrNotFound
	}
	if !IsParent(ctx, accountId, rows[0].ChildId) {
		return ErrForbidden
	}
	return gormdb.GormDB.WithContext(ctx).Exec("delete from child_episode_entries where id = ?", id).Error
}

// ---- medications ----

const maxDoses = 120

func AddMedication(ctx context.Context, accountId string, req hp.MedicationRequest) (hp.Medication, error) {
	if !IsParent(ctx, accountId, req.ChildId) {
		return hp.Medication{}, ErrForbidden
	}
	name := clip(req.Name, 80)
	if name == "" || req.IntervalHours < 0 || req.IntervalHours > 48 || req.MinGapHours < 0 || req.MinGapHours > 48 || req.Days < 0 || req.Days > 30 {
		return hp.Medication{}, ErrInvalid
	}
	if req.IntervalHours > 0 && req.Days == 0 {
		return hp.Medication{}, ErrInvalid
	}
	start, err := parseTime(req.StartTime, time.Now())
	if err != nil {
		return hp.Medication{}, ErrInvalid
	}
	episodeId := req.EpisodeId
	if episodeId == "" {
		if eps, _ := gormdb.CustomSelect[hp.Episode](ctx, "select * from child_episodes where child_id = ? and end_time is null limit 1", req.ChildId); len(eps) > 0 {
			episodeId = eps[0].Id
		}
	}
	m := hp.Medication{Id: uuid.NewString(), ChildId: req.ChildId, EpisodeId: episodeId, Name: name, Dose: clip(req.Dose, 60), IntervalHours: req.IntervalHours,
		MinGapHours: req.MinGapHours, StartTime: start, Active: true, CreatedBy: accountId, CreateTime: time.Now()}
	if req.IntervalHours > 0 {
		end := start.Add(time.Duration(req.Days) * 24 * time.Hour)
		m.EndTime = &end
	}
	if err := gormdb.Create(ctx, &m); err != nil {
		return m, err
	}
	if req.IntervalHours > 0 {
		count := 0
		for t := start; t.Before(*m.EndTime) && count < maxDoses; t = t.Add(time.Duration(req.IntervalHours) * time.Hour) {
			d := hp.Dose{Id: uuid.NewString(), MedicationId: m.Id, ChildId: m.ChildId, DueTime: t, Status: hp.DosePending}
			if err := gormdb.Create(ctx, &d); err != nil {
				return m, err
			}
			count++
		}
	}
	return m, nil
}

func StopMedication(ctx context.Context, accountId, id string) error {
	rows, err := gormdb.CustomSelect[hp.Medication](ctx, "select * from child_medications where id = ?", id)
	if err != nil || len(rows) == 0 {
		return ErrNotFound
	}
	if !IsParent(ctx, accountId, rows[0].ChildId) {
		return ErrForbidden
	}
	db := gormdb.GormDB.WithContext(ctx)
	db.Exec("delete from child_doses where medication_id = ? and status = 'pending'", id)
	return db.Exec("update child_medications set active = false where id = ?", id).Error
}

func loadDose(ctx context.Context, accountId, id string) (hp.Dose, hp.Medication, error) {
	ds, err := gormdb.CustomSelect[hp.Dose](ctx, "select * from child_doses where id = ?", id)
	if err != nil || len(ds) == 0 {
		return hp.Dose{}, hp.Medication{}, ErrNotFound
	}
	if !IsParent(ctx, accountId, ds[0].ChildId) {
		return hp.Dose{}, hp.Medication{}, ErrForbidden
	}
	ms, _ := gormdb.CustomSelect[hp.Medication](ctx, "select * from child_medications where id = ?", ds[0].MedicationId)
	if len(ms) == 0 {
		return hp.Dose{}, hp.Medication{}, ErrNotFound
	}
	return ds[0], ms[0], nil
}

// GiveDose marks a pending dose "podano" -- atomically, so two parents tapping at once record it once -- and tells the other parent.
func GiveDose(ctx context.Context, accountId, id string) (hp.Dose, error) {
	d, m, err := loadDose(ctx, accountId, id)
	if err != nil {
		return d, err
	}
	now := time.Now()
	name := accountName(ctx, accountId)
	res := gormdb.GormDB.WithContext(ctx).Exec("update child_doses set status = 'given', given_time = ?, given_by = ?, given_by_name = ? where id = ? and status = 'pending'", now, accountId, name, id)
	if res.Error != nil {
		return d, res.Error
	}
	if res.RowsAffected == 1 {
		afterGiven(ctx, m, accountId, name, now)
	}
	d.Status, d.GivenTime, d.GivenBy, d.GivenByName = hp.DoseGiven, &now, accountId, name
	return d, nil
}

func SkipDose(ctx context.Context, accountId, id string) error {
	_, _, err := loadDose(ctx, accountId, id)
	if err != nil {
		return err
	}
	return gormdb.GormDB.WithContext(ctx).Exec("update child_doses set status = 'skipped' where id = ? and status = 'pending'", id).Error
}

// GiveAdHoc records an unscheduled dose given right now (medicine "as needed", or an extra dose).
func GiveAdHoc(ctx context.Context, accountId, medicationId string) (hp.Dose, error) {
	ms, err := gormdb.CustomSelect[hp.Medication](ctx, "select * from child_medications where id = ?", medicationId)
	if err != nil || len(ms) == 0 {
		return hp.Dose{}, ErrNotFound
	}
	m := ms[0]
	if !IsParent(ctx, accountId, m.ChildId) {
		return hp.Dose{}, ErrForbidden
	}
	now := time.Now()
	name := accountName(ctx, accountId)
	d := hp.Dose{Id: uuid.NewString(), MedicationId: m.Id, ChildId: m.ChildId, DueTime: now, Status: hp.DoseGiven, GivenTime: &now, GivenBy: accountId, GivenByName: name, AdHoc: true}
	if err := gormdb.Create(ctx, &d); err != nil {
		return d, err
	}
	afterGiven(ctx, m, accountId, name, now)
	return d, nil
}

// afterGiven writes the diary line and notifies the other parent(s).
func afterGiven(ctx context.Context, m hp.Medication, accountId, name string, at time.Time) {
	label := strings.TrimSpace(m.Name + " " + m.Dose)
	if m.EpisodeId != "" {
		e := hp.Entry{Id: uuid.NewString(), EpisodeId: m.EpisodeId, ChildId: m.ChildId, Kind: hp.KindMedication, Text: label, EntryTime: at, CreatedBy: accountId, CreatedByName: name}
		_ = gormdb.Create(ctx, &e)
	}
	notifyOthers(ctx, m.ChildId, accountId, childName(ctx, m.ChildId)+": podano lek", fmt.Sprintf("%s podał(a) %s o %s", name, label, at.In(warsaw).Format("15:04")), "dose_given")
}

func notifyOthers(ctx context.Context, childId, me, title, body, kind string) {
	targets := others(parentIds(ctx, childId), me)
	if len(targets) == 0 {
		return
	}
	go push.SendToAccounts(context.WithoutCancel(ctx), targets, title, body, map[string]string{"type": kind, "child_id": childId})
}

var warsaw = func() *time.Location {
	if l, err := time.LoadLocation("Europe/Warsaw"); err == nil {
		return l
	}
	return time.UTC
}()
