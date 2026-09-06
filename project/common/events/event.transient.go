package eventspg

type EventTransient struct {
	Previous Event
	Current  Event
	Future   Event
}

func (t EventTransient) Update() EventTransient {
	return EventTransient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   Event{},
	}
}
