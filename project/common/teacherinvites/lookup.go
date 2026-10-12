package teacherinvitespg

// LookupRequest is POST /private/teacher-invites/lookup's body --
// authenticated, rate-limited, does not consume the invite.
type LookupRequest struct {
	Pin string `json:"pin,omitempty"`
}

// LookupResponse never errors on an invalid/expired/used PIN -- Usable=false
// + a Reason is easier for a client to render nicely than a raw HTTP
// error. InternalNote is deliberately never included here -- it's
// staff-only (see SelectResponse).
type LookupResponse struct {
	OrganizationName  string `json:"organization_name,omitempty"`
	GroupName         string `json:"group_name,omitempty"`
	AccountType       string `json:"account_type,omitempty"`
	PublicDescription string `json:"public_description,omitempty"`
	Usable            bool   `json:"usable"`
	Reason            string `json:"reason,omitempty"`
}
