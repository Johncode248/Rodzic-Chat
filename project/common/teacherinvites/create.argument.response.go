package teacherinvitespg

type CreateResponse struct {
	TeacherInvite TeacherInvite `json:"teacher_invite,omitempty"`
	// Pin is the plaintext 8-digit PIN -- returned exactly once, here.
	Pin string `json:"pin,omitempty"`
}
