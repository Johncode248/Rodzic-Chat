package chatspg

// MessagePinRequest pins (Pinned: true) or unpins a message. A pinned message is shown at the top
// of its channel; several messages can be pinned at once, newest pin first.
type MessagePinRequest struct {
	MessageId string `json:"message_id,omitempty"`
	Pinned    bool   `json:"pinned"`
}

type MessagePinResponse struct {
	Message Message `json:"message,omitempty"`
}
