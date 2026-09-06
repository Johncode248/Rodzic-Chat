package passwordresetspg

type Transient struct {
	Previous PasswordReset
	Current  PasswordReset
	Future   PasswordReset
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   PasswordReset{},
	}
}
