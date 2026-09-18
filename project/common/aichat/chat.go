package aichatpg

// ChatMessage is one turn in the conversation. Role is "user" or
// "assistant" -- the client owns and resends the whole thread each call
// (no server-side conversation storage, deliberately simple for v1).
type ChatMessage struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// SendRequest: Messages is the full conversation so far (client-managed),
// ending with the new user turn. ChildId is optional -- when set and the
// caller actually is that child's parent, the assistant is given the
// child's name/age as light context; otherwise it answers generically.
type SendRequest struct {
	Messages []ChatMessage `json:"messages,omitempty"`
	ChildId  string        `json:"child_id,omitempty"`
	// Latitude/Longitude, when the client has a location (device GPS or the Rybnik fallback --
	// see LocationManager on the iOS side), seed a default województwo for the NFZ
	// find-a-doctor tool so the assistant doesn't need to ask "w jakim mieście?" before it can
	// search -- the assistant still switches province on its own if the parent names a
	// different city. Optional: nil/0,0 just means no location-based default is offered.
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

// SendResponse: Reply is the assistant's new turn. No usage/limit fields -- this endpoint is
// hard-disabled server-side for MVP (see project/service/transport/aichat.Send), since
// ParentSubscription no longer tracks AI usage at all (project/common/parentsubscriptions was
// stripped down to just Plan).
type SendResponse struct {
	Reply string `json:"reply,omitempty"`
}
