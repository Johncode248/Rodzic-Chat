package organizationspg

type SelectResponse struct {
	Organizations []Organization `json:"organizations,omitempty"`
}
