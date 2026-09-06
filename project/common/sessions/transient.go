package sessionspg

type Transient struct {
	Previous Session
	Current  Session
	Future   Session
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Session{},
	}
}
