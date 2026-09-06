package parentchildpg

type CreateRequest struct {
	ParentId  string `json:"parent_id,omitempty"`
	ChildId   string `json:"child_id,omitempty"`
	Relation  string `json:"relation,omitempty"`
	CanPickup bool   `json:"can_pickup,omitempty"`
}
