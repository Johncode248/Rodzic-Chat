package guardianclaimspg

type Transient struct {
	Previous GuardianClaim
	Current  GuardianClaim
	Future   GuardianClaim
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   GuardianClaim{},
	}
}
