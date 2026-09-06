package absencespg

type DeleteResponse struct {
	Absences []Absence `json:"absences,omitempty"`
}
