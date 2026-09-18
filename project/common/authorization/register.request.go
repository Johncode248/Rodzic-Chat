package authorizationpg

type RegisterRequest struct {
	Name                      string `json:"name,omitempty"`
	Email                     string `json:"email,omitempty"`
	Phone                     string `json:"phone,omitempty"`
	Password                  string `json:"password,omitempty"`
	OrganizationName          string `json:"organization_name,omitempty"`
	OrganizationType          string `json:"organization_type,omitempty"`
	OrganizationChildrenLimit int    `json:"organization_children_limit,omitempty"`
}
