package registrationlinkspg

// PublicChildOption is one child already in the invite's group, offered as
// a pick instead of retyping a name -- see LookupResponse.Children.
type PublicChildOption struct {
	Id        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LookupRequest is POST /private/registration-invites/lookup's body --
// authenticated (any logged-in parent), rate-limited, does not consume the
// invite.
type LookupRequest struct {
	Pin string `json:"pin,omitempty"`
}

// LookupResponse never errors on an invalid/expired/exhausted PIN --
// Usable=false + a Reason is easier for a client to render nicely than a
// raw HTTP error. InternalNote is deliberately never included here -- it's
// staff-only (see SelectResponse).
type LookupResponse struct {
	OrganizationName  string `json:"organization_name,omitempty"`
	GroupName         string `json:"group_name,omitempty"`
	PublicDescription string `json:"public_description,omitempty"`
	Usable            bool   `json:"usable"`
	Reason            string `json:"reason,omitempty"`
	// Children already in this group -- staff never pre-create children
	// (see registrationlinks.Confirm's doc comment), so this list only
	// ever grows from earlier parents confirming through this same invite
	// (or another invite for the same group). Confirming with one of
	// these ids (ConfirmRequest.ChildId) attaches as a second parent to
	// that exact child instead of typing the name again and risking a
	// near-duplicate; the free-text fields remain the way to register a
	// child nobody has added yet. Empty/omitted when the invite isn't
	// usable.
	Children []PublicChildOption `json:"children,omitempty"`
}
