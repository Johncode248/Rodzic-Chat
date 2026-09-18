package childrenpg

// DeleteCondition is empty: "deleting" a child (project/service/internal/children.Delete) never
// touches parent_child rows at all -- it only unenrolls the child from its placówka (clears
// OrganizationId/GroupId), it never removes the Child row itself.
type DeleteCondition struct{}
