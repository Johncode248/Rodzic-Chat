package coparentspg

// CoParent is another parent linked to at least one of the caller's children ("drugi rodzic").
type CoParent struct {
	Id               string   `json:"id,omitempty"`
	Name             string   `json:"name,omitempty"`
	Email            string   `json:"email,omitempty"`
	SharedChildIds   []string `json:"shared_child_ids,omitempty"`
	SharedChildNames []string `json:"shared_child_names,omitempty"`
}

type ListResponse struct {
	CoParents []CoParent `json:"co_parents"`
}

type DisconnectResponse struct {
	UnlinkedChildIds []string `json:"unlinked_child_ids"`
}
