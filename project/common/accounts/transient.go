package accountspg

type Transient struct {
	Previous Account
	Current  Account
	Future   Account
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Account{},
	}
}
