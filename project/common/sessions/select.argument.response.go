package sessionspg

type SelectResponse struct {
	Sessions []Session `json:"sessions,omitempty"`
}
