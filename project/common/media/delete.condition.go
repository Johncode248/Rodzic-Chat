package mediapg

// DeleteCondition is empty: the R2 object deletion is orchestrated by the
// transport layer (which needs the ObjectKey before the DB row disappears),
// not by this internal mutation.
type DeleteCondition struct{}
