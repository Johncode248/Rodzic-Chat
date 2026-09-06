package absencespg

type SelectResponse struct {
	Absences []Absence `json:"absences,omitempty"`
}
