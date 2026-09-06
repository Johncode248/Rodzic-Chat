package organizationspg

type Transient struct {
	Previous Organization
	Current  Organization
	Future   Organization
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Organization{},
	}
}
