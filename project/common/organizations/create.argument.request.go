package organizationspg

type CreateRequest struct {
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	Address       string `json:"address,omitempty"`
	Nip           string `json:"nip,omitempty"`
	ChildrenLimit int    `json:"children_limit,omitempty"`
}
