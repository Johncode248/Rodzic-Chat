package accountspg

type CreateRequest struct {
	Name           string `json:"name,omitempty"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone"`
	OrganizationId string `json:"organization_id,omitempty"`
	AccountType    string `json:"account_type,omitempty"`
}
