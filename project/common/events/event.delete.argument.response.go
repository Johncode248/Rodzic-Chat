package eventspg

type EventDeleteResponse struct {
	Events []Event `json:"events,omitempty"`
}
