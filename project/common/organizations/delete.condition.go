package organizationspg

// DeleteCondition is empty for now: organizations do not orchestrate any
// cascading dependent deletes yet (accounts/children/groups belonging to an
// organization are left in place -- future work).
type DeleteCondition struct{}
