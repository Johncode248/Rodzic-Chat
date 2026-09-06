package childrenpg

// DeleteCondition is empty: deleting a child does not currently cascade into
// parent_child relations (out of scope for this task).
type DeleteCondition struct{}
