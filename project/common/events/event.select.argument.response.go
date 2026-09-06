package eventspg

type EventSelectResponse struct {
	Events []Event `json:"events,omitempty"`
}
