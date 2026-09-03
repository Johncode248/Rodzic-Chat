package databasepg

type Group struct {
	Filters []Filter `json:"filters,omitempty"`
}
