package mediapg

type Transient struct {
	Previous Media
	Current  Media
	Future   Media
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Media{},
	}
}
