package organizationspg

type SubscriptionStatus int

const (
	SubscriptionStatusUnknown SubscriptionStatus = iota
	SubscriptionStatusTrial
	SubscriptionStatusActive
	SubscriptionStatusPastDue
	SubscriptionStatusCancelled
)

func (s SubscriptionStatus) String() string {
	switch s {
	case SubscriptionStatusTrial:
		return "Trial"
	case SubscriptionStatusActive:
		return "Active"
	case SubscriptionStatusPastDue:
		return "PastDue"
	case SubscriptionStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

func (s SubscriptionStatus) Ordinal() int {
	return int(s)
}
