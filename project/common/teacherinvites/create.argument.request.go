package teacherinvitespg

type CreateRequest struct {
	GroupId     string `json:"group_id,omitempty"`
	AccountType string `json:"account_type,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
}
