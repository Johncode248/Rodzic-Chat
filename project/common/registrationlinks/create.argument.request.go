package registrationlinkspg

type CreateRequest struct {
	GroupId       string `json:"group_id,omitempty"`
	ChildrenCount int    `json:"children_count,omitempty"`
}
