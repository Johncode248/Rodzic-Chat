package eventspg

type RegistrationTransient struct {
	Previous EventRegistration
	Current  EventRegistration
	Future   EventRegistration
}

func (t RegistrationTransient) Update() RegistrationTransient {
	return RegistrationTransient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   EventRegistration{},
	}
}
