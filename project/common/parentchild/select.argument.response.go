package parentchildpg

type SelectResponse struct {
	ParentChildren []ParentChild `json:"parent_children,omitempty"`
}
