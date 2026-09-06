package accountorganizationspg

type Transient struct {
	Previous AccountOrganization
	Current  AccountOrganization
	Future   AccountOrganization
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   AccountOrganization{},
	}
}
