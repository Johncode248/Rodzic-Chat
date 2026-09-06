package absencespg

type Transient struct {
	Previous Absence
	Current  Absence
	Future   Absence
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Absence{},
	}
}
