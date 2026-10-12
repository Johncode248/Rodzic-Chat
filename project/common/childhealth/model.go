// Package childhealthpg holds a child's health card (blood type, allergies, doctors, measurements) and the illness
// diary (episodes, log entries, medication schedules and their doses). Everything here is visible ONLY to the parents
// linked to the child (parent_children); no vaccination data, no PESEL, no documents.
package childhealthpg

import "time"

// Profile is the health card header of a child. It is not a table of its own: blood type and note are columns of
// children (blood_type, health_note).
type Profile struct {
	ChildId   string `json:"child_id"`
	BloodType string `json:"blood_type"` // "", "0+", "0-", "A+", "A-", "B+", "B-", "AB+", "AB-"
	Note      string `json:"note"`
}

type Allergy struct {
	Id         string    `json:"id" gorm:"primaryKey"`
	ChildId    string    `json:"child_id" gorm:"not null;index"`
	Name       string    `json:"name" gorm:"not null"`
	Severity   string    `json:"severity"` // mild | moderate | severe
	Reaction   string    `json:"reaction"`
	CreateTime time.Time `json:"create_time"`
}

type Doctor struct {
	Id         string    `json:"id" gorm:"primaryKey"`
	ChildId    string    `json:"child_id" gorm:"not null;index"`
	Name       string    `json:"name" gorm:"not null"`
	Specialty  string    `json:"specialty"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	CreateTime time.Time `json:"create_time"`
}

type Measurement struct {
	Id         string    `json:"id" gorm:"primaryKey"`
	ChildId    string    `json:"child_id" gorm:"not null;index"`
	Date       time.Time `json:"date" gorm:"not null"`
	WeightKg   float64   `json:"weight_kg"` // 0 = not measured
	HeightCm   float64   `json:"height_cm"` // 0 = not measured
	CreateTime time.Time `json:"create_time"`
}

// Episode is one illness ("kaszel i gorączka"), open until EndTime is set.
type Episode struct {
	Id         string     `json:"id" gorm:"primaryKey"`
	ChildId    string     `json:"child_id" gorm:"not null;index"`
	Title      string     `json:"title" gorm:"not null"`
	StartTime  time.Time  `json:"start_time" gorm:"not null"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	CreatedBy  string     `json:"created_by"`
	CreateTime time.Time  `json:"create_time"`
}

const (
	KindTemperature = "temperature"
	KindSymptom     = "symptom"
	KindNote        = "note"
	KindMedication  = "medication" // written automatically when a dose is marked given
)

// Entry is a line of the diary.
type Entry struct {
	Id            string    `json:"id" gorm:"primaryKey"`
	EpisodeId     string    `json:"episode_id" gorm:"not null;index"`
	ChildId       string    `json:"child_id" gorm:"not null;index"`
	Kind          string    `json:"kind" gorm:"not null"`
	Text          string    `json:"text"`
	Value         float64   `json:"value"` // temperature in °C for KindTemperature
	EntryTime     time.Time `json:"entry_time" gorm:"not null"`
	CreatedBy     string    `json:"created_by"`
	CreatedByName string    `json:"created_by_name"`
}

// Medication is a schedule the parent entered from the leaflet / the doctor's instructions -- the app never picks doses.
type Medication struct {
	Id            string     `json:"id" gorm:"primaryKey"`
	ChildId       string     `json:"child_id" gorm:"not null;index"`
	EpisodeId     string     `json:"episode_id" gorm:"index"`
	Name          string     `json:"name" gorm:"not null"`
	Dose          string     `json:"dose"`
	IntervalHours int        `json:"interval_hours"` // 0 = as needed (no schedule)
	MinGapHours   int        `json:"min_gap_hours"`  // shows a warning when the last given dose is younger
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Active        bool       `json:"active" gorm:"not null;default:true"`
	CreatedBy     string     `json:"created_by"`
	CreateTime    time.Time  `json:"create_time"`
}

const (
	DosePending = "pending"
	DoseGiven   = "given"
	DoseSkipped = "skipped"
)

type Dose struct {
	Id           string     `json:"id" gorm:"primaryKey"`
	MedicationId string     `json:"medication_id" gorm:"not null;index"`
	ChildId      string     `json:"child_id" gorm:"not null;index"`
	DueTime      time.Time  `json:"due_time" gorm:"not null;index"`
	Status       string     `json:"status" gorm:"not null;default:pending"`
	GivenTime    *time.Time `json:"given_time,omitempty"`
	GivenBy      string     `json:"given_by,omitempty"`
	GivenByName  string     `json:"given_by_name,omitempty"`
	RemindCount  int        `json:"-" gorm:"not null;default:0"`
	AdHoc        bool       `json:"ad_hoc"`
}

// ---- responses ----

type Overview struct {
	Profile      Profile       `json:"profile"`
	Allergies    []Allergy     `json:"allergies"`
	Doctors      []Doctor      `json:"doctors"`
	Measurements []Measurement `json:"measurements"`
	ActiveTitle  string        `json:"active_title,omitempty"` // title of the open illness, "" when healthy
	NextDose     *NextDose     `json:"next_dose,omitempty"`
}

type NextDose struct {
	DoseId  string    `json:"dose_id"`
	Name    string    `json:"name"`
	Dose    string    `json:"dose"`
	DueTime time.Time `json:"due_time"`
	Overdue bool      `json:"overdue"`
}

type MedicationView struct {
	Medication
	Doses []Dose `json:"doses"`
	// LastGiven is when a dose of this medication was last given (for the "za wcześnie?" warning).
	LastGiven *time.Time `json:"last_given,omitempty"`
}

type Current struct {
	Episode     Episode          `json:"episode"`
	Entries     []Entry          `json:"entries"`
	Medications []MedicationView `json:"medications"`
}

type IllnessOverview struct {
	Current *Current  `json:"current,omitempty"`
	Past    []Episode `json:"past"`
}

type EpisodeDetail struct {
	Episode Episode `json:"episode"`
	Entries []Entry `json:"entries"`
}

// ---- requests ----

type ProfileRequest struct {
	ChildId   string `json:"child_id"`
	BloodType string `json:"blood_type"`
	Note      string `json:"note"`
}

type AllergyRequest struct {
	ChildId  string `json:"child_id"`
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Reaction string `json:"reaction"`
}

type DoctorRequest struct {
	ChildId   string `json:"child_id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}

type MeasurementRequest struct {
	ChildId  string  `json:"child_id"`
	Date     string  `json:"date"` // RFC3339 or YYYY-MM-DD
	WeightKg float64 `json:"weight_kg"`
	HeightCm float64 `json:"height_cm"`
}

type EpisodeRequest struct {
	ChildId string `json:"child_id"`
	Title   string `json:"title"`
}

type EntryRequest struct {
	EpisodeId string  `json:"episode_id"`
	Kind      string  `json:"kind"`
	Text      string  `json:"text"`
	Value     float64 `json:"value"`
	EntryTime string  `json:"entry_time"` // RFC3339, "" = now
}

type MedicationRequest struct {
	ChildId       string `json:"child_id"`
	EpisodeId     string `json:"episode_id"`
	Name          string `json:"name"`
	Dose          string `json:"dose"`
	IntervalHours int    `json:"interval_hours"`
	MinGapHours   int    `json:"min_gap_hours"`
	Days          int    `json:"days"`       // schedule length when IntervalHours > 0
	StartTime     string `json:"start_time"` // RFC3339, "" = now
}

func (Allergy) TableName() string     { return "child_allergies" }
func (Doctor) TableName() string      { return "child_doctors" }
func (Measurement) TableName() string { return "child_measurements" }
func (Episode) TableName() string     { return "child_episodes" }
func (Entry) TableName() string       { return "child_episode_entries" }
func (Medication) TableName() string  { return "child_medications" }
func (Dose) TableName() string        { return "child_doses" }
