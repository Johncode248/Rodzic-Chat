// Package reminders is the backend's first (and, as of this package, only) scheduled/background
// job: once a day it checks for payments and event-registration deadlines landing exactly
// reminderLeadTime from now and pushes a one-time notification to the affected parent(s). See
// this package's SendDueReminders doc comment for why "exactly N days out, checked once daily"
// needs no extra idempotency bookkeeping (no "already sent" flag/table) to avoid duplicate pushes.
package reminders

import (
	"context"
	"fmt"
	"github.com/Johncode248/Rodzic-Chat/project/common/modules"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/push"
)

// reminderLeadTime is how far ahead of a deadline the one reminder goes out (product direction:
// "3 dni przed terminem, jedno przypomnienie" -- no second/repeated reminder).
const reminderLeadTime = 3 * 24 * time.Hour

// SendDueReminders pushes payment-due and event-RSVP-deadline reminders for whatever falls
// exactly reminderLeadTime from now, truncated to whole days -- e.g. run today, it targets
// everything due/closing on today+3's calendar date. Because this only ever matches deadlines
// landing in that one specific future day, and the caller (see runScheduler in
// project/service/internal/module.go) only invokes this once per calendar day, no payment or
// event can ever be matched by two different daily runs -- that's the entire "send exactly once"
// guarantee, without a dedicated "reminder already sent" column/table.
func SendDueReminders(ctx context.Context, now time.Time) {
	if !modules.Organizations {
		return // payment and school-event reminders need the placówka module
	}
	targetDayStart := now.Add(reminderLeadTime).Truncate(24 * time.Hour)
	targetDayEnd := targetDayStart.AddDate(0, 0, 1)

	sendPaymentReminders(ctx, targetDayStart, targetDayEnd)
	sendEventReminders(ctx, targetDayStart, targetDayEnd)
}

func sendPaymentReminders(ctx context.Context, dayStart, dayEnd time.Time) {
	result, err := payments.Select(ctx, databasepg.Query{
		Limit: 1000,
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{
			{Key: "status", Operator: "=", Value: paymentspg.PaymentStatusUnpaid.Ordinal()},
			{Key: "due_date", Operator: ">=", Value: dayStart},
			{Key: "due_date", Operator: "<", Value: dayEnd},
		}}},
	})
	if err != nil {
		return
	}

	for _, p := range result.Payments {
		if p.ParentId == "" {
			continue
		}
		body := fmt.Sprintf("%s -- termin za 3 dni (%s)", p.Title, dayStart.Format("02.01.2006"))
		push.SendToAccounts(ctx, []string{p.ParentId}, "Zbliża się termin płatności", body, map[string]string{
			"type":       "payment_due",
			"payment_id": p.Id,
		})
	}
}

func sendEventReminders(ctx context.Context, dayStart, dayEnd time.Time) {
	eventsResult, err := events.SelectEvents(ctx, eventspg.EventSelectArgument{
		Query: databasepg.Query{
			Limit: 500,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "registration_deadline", Operator: ">=", Value: dayStart},
				{Key: "registration_deadline", Operator: "<", Value: dayEnd},
			}}},
		},
	})
	if err != nil || len(eventsResult.Events) == 0 {
		return
	}

	for _, event := range eventsResult.Events {
		if event.GroupId == "" {
			continue
		}
		remindMissingRegistrations(ctx, event)
	}
}

// remindMissingRegistrations pushes to the parent(s) of every child in event's group that has no
// EventRegistration row at all yet (a Declined/Confirmed row means the parent already acted --
// only a genuinely missing row still needs a nudge).
func remindMissingRegistrations(ctx context.Context, event eventspg.Event) {
	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Limit:  500,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: event.GroupId}}}},
		},
	})
	if err != nil || len(childrenResult.Children) == 0 {
		return
	}

	regsResult, err := events.SelectRegistrations(ctx, eventspg.RegistrationSelectArgument{
		Query: databasepg.Query{
			Limit:  500,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "event_id", Operator: "=", Value: event.Id}}}},
		},
	})
	if err != nil {
		return
	}
	registered := make(map[string]bool, len(regsResult.EventRegistrations))
	for _, r := range regsResult.EventRegistrations {
		registered[r.ChildId] = true
	}

	var missingChildIds []string
	for _, child := range childrenResult.Children {
		if !registered[child.Id] {
			missingChildIds = append(missingChildIds, child.Id)
		}
	}
	if len(missingChildIds) == 0 {
		return
	}

	linksResult, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Limit:  500,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "child_id", Operator: "IN", Value: missingChildIds}}}},
		},
	})
	if err != nil || len(linksResult.ParentChildren) == 0 {
		return
	}

	parentIds := make([]string, 0, len(linksResult.ParentChildren))
	seen := map[string]bool{}
	for _, link := range linksResult.ParentChildren {
		if !seen[link.ParentId] {
			seen[link.ParentId] = true
			parentIds = append(parentIds, link.ParentId)
		}
	}

	body := fmt.Sprintf("%s -- zgłoszenia do %s", event.Title, event.RegistrationDeadline.Format("02.01.2006"))
	push.SendToAccounts(ctx, parentIds, "Zbliża się termin zgłoszenia", body, map[string]string{
		"type":     "event_registration_due",
		"event_id": event.Id,
	})
}
