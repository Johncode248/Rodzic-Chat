package childrenpg

type Transient struct {
	Previous Child
	Current  Child
	Future   Child
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Child{},
	}
}
