package childevents

import (
	"time"

	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
)

// Span is one concrete occurrence of an event.
type Span struct {
	Start time.Time
	End   *time.Time
}

const maxOccurrences = 1000

// Occurrences expands an event into its occurrences overlapping [from, to]. A non-repeating event
// yields exactly one span (regardless of the window, so a past or far-future one-off is never
// silently hidden); a repeating one yields every occurrence in the window, until RecurrenceUntil.
// Every occurrence keeps the original duration (multi-day events stay multi-day).
func Occurrences(event childeventspg.ChildEvent, from, to time.Time) []Span {
	// Work in the zone the event was scheduled in (see ChildEvent.TimeZone).
	if loc, err := time.LoadLocation(event.TimeZone); err == nil && event.TimeZone != "" {
		event.StartTime = event.StartTime.In(loc)
		if event.EndTime != nil {
			end := event.EndTime.In(loc)
			event.EndTime = &end
		}
	}
	var duration time.Duration
	if event.EndTime != nil {
		duration = event.EndTime.Sub(event.StartTime)
	}
	span := func(start time.Time, k int) Span {
		s := Span{Start: start}
		if event.EndTime != nil {
			// The end shifts exactly like the start (same wall-clock time even across a DST change);
			// a fixed +duration would drift by an hour over a DST boundary. If clamping makes the
			// shifted end fall before the start, fall back to the original duration.
			end := shift(*event.EndTime, event.Recurrence, k)
			if k == 0 || end.Before(start) {
				end = start.Add(duration)
				if k == 0 {
					end = *event.EndTime
				}
			}
			s.End = &end
		}
		return s
	}

	if event.Recurrence == childeventspg.RecurrenceNone {
		return []Span{span(event.StartTime, 0)}
	}

	interval := event.RecurrenceInterval
	if interval < 1 {
		interval = 1
	}
	var spans []Span
	for k := 0; k < maxOccurrences; k++ {
		start := shift(event.StartTime, event.Recurrence, k*interval)
		if start.After(to) || (event.RecurrenceUntil != nil && start.After(*event.RecurrenceUntil)) {
			break
		}
		occ := span(start, k*interval)
		end := occ.Start
		if occ.End != nil {
			end = *occ.End
		}
		if !end.Before(from) {
			spans = append(spans, occ)
		}
	}
	return spans
}

// shift moves t forward by n units of freq, always from the ORIGINAL start (so a monthly event on
// the 31st lands on the last day of shorter months instead of drifting into the next month).
func shift(t time.Time, freq string, n int) time.Time {
	switch freq {
	case childeventspg.RecurrenceDaily:
		return t.AddDate(0, 0, n)
	case childeventspg.RecurrenceWeekly:
		return t.AddDate(0, 0, 7*n)
	case childeventspg.RecurrenceMonthly:
		return addMonthsClamped(t, n)
	case childeventspg.RecurrenceYearly:
		return addMonthsClamped(t, 12*n)
	}
	return t
}

func addMonthsClamped(t time.Time, months int) time.Time {
	year, month, day := t.Date()
	target := time.Date(year, month, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()).AddDate(0, months, 0)
	lastDay := time.Date(target.Year(), target.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(target.Year(), target.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// validZone returns name when it is a loadable IANA zone, else "" (the offset already stored in
// the timestamps is then used as-is).
func validZone(name string) string {
	if name == "" {
		return ""
	}
	if _, err := time.LoadLocation(name); err != nil {
		return ""
	}
	return name
}
