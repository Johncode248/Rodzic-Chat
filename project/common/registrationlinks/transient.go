package registrationlinkspg

type Transient struct {
	Previous RegistrationLink
	Current  RegistrationLink
	Future   RegistrationLink
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   RegistrationLink{},
	}
}
