package groupspg

// DeleteCondition is empty: deleting a group does not currently cascade into
// children (a child's group_id is left dangling -- future work).
type DeleteCondition struct{}
