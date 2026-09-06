package groupspg

type Transient struct {
	Previous Group
	Current  Group
	Future   Group
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Group{},
	}
}
