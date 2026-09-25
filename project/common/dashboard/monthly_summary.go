package dashboardpg

import "time"

// MonthlySummaryTodoPayment is one still-unpaid Payment (project/common/payments) whose DueDate
// falls in the requested month -- shown on the parent's "Panel główny" as something to act on.
type MonthlySummaryTodoPayment struct {
	Id           string     `json:"id,omitempty"`
	Title        string     `json:"title,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	AmountGrosze int64      `json:"amount_grosze,omitempty"`
}

// MonthlySummaryTodoEvent is one (Event, Child) pair -- project/common/events -- where that
// child has no EventRegistration row at all yet, and the event's RegistrationDeadline falls in
// the requested month.
type MonthlySummaryTodoEvent struct {
	Id                   string     `json:"id,omitempty"`
	Title                string     `json:"title,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	ChildId              string     `json:"child_id,omitempty"`
}

// MonthlySummaryTodoConsent is one open poll or consent request (project/common/chats) -- posted
// in the requested month, in a channel the parent belongs to -- that the parent hasn't voted on
// / confirmed yet. Kind is "poll" or "consent".
type MonthlySummaryTodoConsent struct {
	MessageId string `json:"message_id,omitempty"`
	ChannelId string `json:"channel_id,omitempty"`
	Content   string `json:"content,omitempty"`
	Kind      string `json:"kind,omitempty"`
}

// MonthlySummaryResponse is a parent's "co do zrobienia / co już zrobione" digest for one
// calendar month (Month is 1-indexed, 1=styczeń), fanned out across payments/absences/meals/
// events/chats at request time -- see dashboardpg.FeedResponse's doc comment for the same
// "no dedicated table" precedent this follows. The "done" side is deliberately just counts/sums
// (matching the product's own "2 nieobecności, 15 obiadów, 450 zł opłacone" framing) while the
// "todo" side carries enough per-item detail (title, amount/deadline) for the parent to act on
// without a second screen.
type MonthlySummaryResponse struct {
	Year  int `json:"year,omitempty"`
	Month int `json:"month,omitempty"`

	TodoPayments []MonthlySummaryTodoPayment `json:"todo_payments,omitempty"`
	TodoEvents   []MonthlySummaryTodoEvent   `json:"todo_events,omitempty"`
	TodoConsents []MonthlySummaryTodoConsent `json:"todo_consents,omitempty"`

	AbsencesCount          int   `json:"absences_count,omitempty"`
	MealCancellationsCount int   `json:"meal_cancellations_count,omitempty"`
	PaidAmountGrosze       int64 `json:"paid_amount_grosze,omitempty"`
	ConfirmedEventsCount   int   `json:"confirmed_events_count,omitempty"`
}
