package teacherinvitespg

type PublicInfoResponse struct {
	OrganizationName string `json:"organization_name,omitempty"`
	GroupName        string `json:"group_name,omitempty"`
	Usable           bool   `json:"usable"`
	Reason           string `json:"reason,omitempty"`
}
