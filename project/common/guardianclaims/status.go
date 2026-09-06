package guardianclaimspg

type GuardianClaimStatus int

const (
	GuardianClaimStatusUnknown GuardianClaimStatus = iota
	GuardianClaimStatusPending
	GuardianClaimStatusApproved
	GuardianClaimStatusRejected
)

func (s GuardianClaimStatus) String() string {
	switch s {
	case GuardianClaimStatusPending:
		return "Pending"
	case GuardianClaimStatusApproved:
		return "Approved"
	case GuardianClaimStatusRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}

func (s GuardianClaimStatus) Ordinal() int {
	return int(s)
}
