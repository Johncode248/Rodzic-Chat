package credentialspg

type Transient struct {
	Previous Credential
	Current  Credential
	Future   Credential
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Credential{},
	}
}
