package featurespg

type Transient struct {
	Previous OrganizationFeature
	Current  OrganizationFeature
	Future   OrganizationFeature
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   OrganizationFeature{},
	}
}
