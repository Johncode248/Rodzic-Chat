package childevents

import (
	"testing"
	"time"

	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
)

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 10, 0, 0, 0, time.UTC) }

func TestOccurrencesOneOffAlwaysReturned(t *testing.T) {
	e := childeventspg.ChildEvent{StartTime: date(2020, 1, 1)}
	if got := Occurrences(e, date(2026, 1, 1), date(2027, 1, 1)); len(got) != 1 {
		t.Fatalf("one-off outside window must still be returned once, got %d", len(got))
	}
}

func TestOccurrencesMonthlyClampsToMonthEnd(t *testing.T) {
	e := childeventspg.ChildEvent{StartTime: date(2026, 1, 31), Recurrence: "monthly", RecurrenceInterval: 1}
	got := Occurrences(e, date(2026, 1, 1), date(2026, 4, 30))
	want := []time.Time{date(2026, 1, 31), date(2026, 2, 28), date(2026, 3, 31), date(2026, 4, 30)}
	if len(got) != len(want) {
		t.Fatalf("got %d occurrences, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Start.Equal(want[i]) {
			t.Errorf("occurrence %d = %v, want %v", i, got[i].Start, want[i])
		}
	}
}

func TestOccurrencesWeeklyUntilAndDuration(t *testing.T) {
	end := date(2026, 3, 3) // 2-day event
	until := date(2026, 3, 20)
	e := childeventspg.ChildEvent{StartTime: date(2026, 3, 1), EndTime: &end, Recurrence: "weekly", RecurrenceInterval: 1, RecurrenceUntil: &until}
	got := Occurrences(e, date(2026, 3, 1), date(2026, 12, 1))
	if len(got) != 3 { // Mar 1, 8, 15 (Mar 22 is after until)
		t.Fatalf("got %d occurrences, want 3", len(got))
	}
	if d := got[1].End.Sub(got[1].Start); d != 48*time.Hour {
		t.Errorf("duration = %v, want 48h", d)
	}
}

func TestOccurrencesKeepWallClockAcrossDST(t *testing.T) {
	warsaw, _ := time.LoadLocation("Europe/Warsaw")
	start := time.Date(2026, 10, 20, 18, 45, 0, 0, warsaw) // CEST
	end := time.Date(2026, 10, 22, 19, 45, 0, 0, warsaw)
	e := childeventspg.ChildEvent{StartTime: start, EndTime: &end, Recurrence: "weekly", RecurrenceInterval: 1, TimeZone: "Europe/Warsaw"}
	got := Occurrences(e, start, start.AddDate(0, 0, 20)) // Oct 20, 27, Nov 3 -- clocks go back on Oct 25
	if len(got) < 2 {
		t.Fatalf("got %d occurrences", len(got))
	}
	for i, o := range got {
		h, m, _ := o.Start.In(warsaw).Clock()
		eh, em, _ := o.End.In(warsaw).Clock()
		if h != 18 || m != 45 || eh != 19 || em != 45 {
			t.Errorf("occurrence %d = %v -> %v (Warsaw wall clock changed)", i, o.Start.In(warsaw), o.End.In(warsaw))
		}
	}
}
