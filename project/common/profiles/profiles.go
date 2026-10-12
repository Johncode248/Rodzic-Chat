package profilespg

// Kind is what a profile IS from the user's point of view; one login holds at most one profile per kind
// (you switch between DIFFERENT kinds, never between two parents).
const (
	KindParent    = "parent"
	KindDirector  = "director"
	KindOrganizer = "organizer"
)

type Profile struct {
	AccountId        string `json:"account_id"`
	Kind             string `json:"kind"`
	Name             string `json:"name"`
	OrganizationName string `json:"organization_name,omitempty"`
	Current          bool   `json:"current"`
}

type ListResponse struct {
	Profiles []Profile `json:"profiles"`
}

type AddRequest struct {
	Kind             string `json:"kind"`
	Name             string `json:"name,omitempty"`
	Phone            string `json:"phone,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
}

type SwitchRequest struct {
	AccountId string `json:"account_id"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
