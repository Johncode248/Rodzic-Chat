package teacherinvitespg

type Transient struct {
	Previous TeacherInvite
	Current  TeacherInvite
	Future   TeacherInvite
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   TeacherInvite{},
	}
}
