package parentchildpg

type Transient struct {
	Previous ParentChild
	Current  ParentChild
	Future   ParentChild
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   ParentChild{},
	}
}
