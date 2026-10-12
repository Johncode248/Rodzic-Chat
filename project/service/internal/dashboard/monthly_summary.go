package dashboard

import (
	"context"
	"github.com/Johncode248/Rodzic-Chat/project/common/modules"
	"sort"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	absencespg "github.com/Johncode248/Rodzic-Chat/project/common/absences"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	dashboardpg "github.com/Johncode248/Rodzic-Chat/project/common/dashboard"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	mealspg "github.com/Johncode248/Rodzic-Chat/project/common/meals"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	paymentspg "github.com/Johncode248/Rodzic-Chat/project/common/payments"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/absences"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/meals"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/payments"
)

// MonthlySummary builds a parent's "Panel główny" monthly digest -- "co do zrobienia" (unpaid
// payments due, events needing an RSVP, open polls/consents) and "co już zrobione" (absences
// reported, meals cancelled, amount paid, events confirmed) -- for one calendar month, fanned
// out across the already-existing payments/absences/meals/events/chats sources at request time,
// same "no dedicated table" precedent as dashboard.Feed. month is 1-indexed (1=styczeń).
func MonthlySummary(ctx context.Context, accountId string, year, month int) (res dashboardpg.MonthlySummaryResponse, err error) {
	res.Year = year
	res.Month = month
	if !modules.Organizations {
		return res, nil // payments, school events, polls and consents exist only with the placówka module
	}

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0)

	childIds, err := childIdsForParent(ctx, accountId)
	if err != nil {
		return res, err
	}

	if err = addPayments(ctx, accountId, monthStart, monthEnd, &res); err != nil {
		return res, err
	}

	if len(childIds) > 0 {
		if err = addAbsencesAndMeals(ctx, childIds, monthStart, monthEnd, &res); err != nil {
			return res, err
		}
	}

	if err = addEvents(ctx, accountId, childIds, monthStart, monthEnd, &res); err != nil {
		return res, err
	}

	if err = addPollsAndConsents(ctx, accountId, monthStart, monthEnd, &res); err != nil {
		return res, err
	}

	sort.Slice(res.TodoPayments, func(i, j int) bool {
		return res.TodoPayments[i].DueDate.Before(*res.TodoPayments[j].DueDate)
	})

	return res, nil
}

// addPayments fills TodoPayments (unpaid, due in [monthStart, monthEnd)) and PaidAmountGrosze
// (paid or reconciled, with whichever of PaidAt/ReconciledAt is set falling in that range) --
// there is no due-date filter on the public GET /private/payments endpoint today, so this fetches
// the parent's full payment list (bounded by the 500 limit below -- plenty for one parent) and
// filters in Go rather than adding one just for this internal caller.
func addPayments(ctx context.Context, accountId string, monthStart, monthEnd time.Time, res *dashboardpg.MonthlySummaryResponse) error {
	result, err := payments.Select(ctx, databasepg.Query{
		Limit:  500,
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "parent_id", Operator: "=", Value: accountId}}}},
	})
	if err != nil {
		return err
	}

	for _, p := range result.Payments {
		if p.Status == paymentspg.PaymentStatusUnpaid && inRange(p.DueDate, monthStart, monthEnd) {
			res.TodoPayments = append(res.TodoPayments, dashboardpg.MonthlySummaryTodoPayment{
				Id: p.Id, Title: p.Title, DueDate: p.DueDate, AmountGrosze: p.BaseAmountGrosze,
			})
		}
		if p.Status == paymentspg.PaymentStatusPaid || p.Status == paymentspg.PaymentStatusReconciled {
			paidAt := p.PaidAt
			if paidAt == nil {
				paidAt = p.ReconciledAt
			}
			if inRange(paidAt, monthStart, monthEnd) {
				res.PaidAmountGrosze += p.BaseAmountGrosze
			}
		}
	}
	return nil
}

// addAbsencesAndMeals fills AbsencesCount/MealCancellationsCount -- both endpoints already
// support the "does [date_from, date_to] overlap [from, to]" filter (see absencesgt's
// dateRangeFilters), reused here directly against the month's own bounds.
func addAbsencesAndMeals(ctx context.Context, childIds []string, monthStart, monthEnd time.Time, res *dashboardpg.MonthlySummaryResponse) error {
	overlapFilters := []databasepg.Filter{
		{Key: "date_from", Operator: "<", Value: monthEnd},
		{Key: "date_to", Operator: ">=", Value: monthStart},
	}

	absencesResult, err := absences.Select(ctx, absencespg.SelectArgument{
		Query: databasepg.Query{
			Limit: 500,
			Groups: []databasepg.Group{{Filters: append(
				[]databasepg.Filter{{Key: "child_id", Operator: "IN", Value: childIds}},
				overlapFilters...,
			)}},
		},
	})
	if err != nil {
		return err
	}
	res.AbsencesCount = len(absencesResult.Absences)

	mealsResult, err := meals.Select(ctx, mealspg.SelectArgument{
		Query: databasepg.Query{
			Limit: 500,
			Groups: []databasepg.Group{{Filters: append(
				[]databasepg.Filter{{Key: "child_id", Operator: "IN", Value: childIds}},
				overlapFilters...,
			)}},
		},
	})
	if err != nil {
		return err
	}
	res.MealCancellationsCount = len(mealsResult.MealCancellations)
	return nil
}

// addEvents fills TodoEvents (one entry per (event, child) with no registration row at all yet,
// for events whose RegistrationDeadline falls in the month) and ConfirmedEventsCount.
func addEvents(ctx context.Context, accountId string, childIds []string, monthStart, monthEnd time.Time, res *dashboardpg.MonthlySummaryResponse) error {
	groupIds, err := chats.ParentGroupIds(ctx, accountId)
	if err != nil || len(groupIds) == 0 {
		return err
	}

	eventsResult, err := events.SelectEvents(ctx, eventspg.EventSelectArgument{
		Query: databasepg.Query{
			Limit:  500,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "group_id", Operator: "IN", Value: groupIds}}}},
		},
	})
	if err != nil {
		return err
	}

	var eventIds []string
	eventById := map[string]eventspg.Event{}
	for _, e := range eventsResult.Events {
		if !inRange(e.RegistrationDeadline, monthStart, monthEnd) {
			continue
		}
		eventIds = append(eventIds, e.Id)
		eventById[e.Id] = e
	}
	if len(eventIds) == 0 || len(childIds) == 0 {
		return nil
	}

	regsResult, err := events.SelectRegistrations(ctx, eventspg.RegistrationSelectArgument{
		Query: databasepg.Query{
			Limit: 1000,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "event_id", Operator: "IN", Value: eventIds},
				{Key: "child_id", Operator: "IN", Value: childIds},
			}}},
		},
	})
	if err != nil {
		return err
	}

	hasRegistration := map[string]bool{}
	for _, r := range regsResult.EventRegistrations {
		hasRegistration[r.EventId+"|"+r.ChildId] = true
		if r.Status == eventspg.RegistrationStatusConfirmed {
			res.ConfirmedEventsCount++
		}
	}
	for _, eventId := range eventIds {
		e := eventById[eventId]
		for _, childId := range childIds {
			if !hasRegistration[eventId+"|"+childId] {
				res.TodoEvents = append(res.TodoEvents, dashboardpg.MonthlySummaryTodoEvent{
					Id: e.Id, Title: e.Title, RegistrationDeadline: e.RegistrationDeadline, ChildId: childId,
				})
			}
		}
	}
	return nil
}

// addPollsAndConsents fills TodoConsents -- every poll/consent message posted in the month, in a
// channel accountId is a member of, that accountId hasn't voted on / confirmed yet. There is no
// existing "polls/consents awaiting me" query anywhere in the codebase (ConsentStatus/
// GetPollDetail are both per-message lookups only -- see their doc comments), so this is the
// first cross-channel version: list member channels, list Poll/Consent messages in them created
// this month, then diff against this account's own votes/confirmations.
func addPollsAndConsents(ctx context.Context, accountId string, monthStart, monthEnd time.Time, res *dashboardpg.MonthlySummaryResponse) error {
	channelIds, err := memberChannelIds(ctx, accountId)
	if err != nil || len(channelIds) == 0 {
		return err
	}

	messagesResult, err := chats.SelectMessages(ctx, chatspg.MessageSelectArgument{
		Query: databasepg.Query{
			Limit: 500,
			Groups: []databasepg.Group{{Filters: []databasepg.Filter{
				{Key: "channel_id", Operator: "IN", Value: channelIds},
				{Key: "type", Operator: "IN", Value: []int{chatspg.MessageTypePoll.Ordinal(), chatspg.MessageTypeConsent.Ordinal()}},
			}}},
		},
	})
	if err != nil {
		return err
	}

	var pollMessageIds, consentMessageIds []string
	messageById := map[string]chatspg.Message{}
	for _, m := range messagesResult.Messages {
		t := m.CreateTime
		if t.Before(monthStart) || !t.Before(monthEnd) {
			continue
		}
		messageById[m.Id] = m
		if m.Type == chatspg.MessageTypePoll {
			pollMessageIds = append(pollMessageIds, m.Id)
		} else {
			consentMessageIds = append(consentMessageIds, m.Id)
		}
	}

	if len(pollMessageIds) > 0 {
		if err := addPendingPolls(ctx, accountId, pollMessageIds, messageById, res); err != nil {
			return err
		}
	}
	if len(consentMessageIds) > 0 {
		if err := addPendingConsents(ctx, accountId, consentMessageIds, messageById, res); err != nil {
			return err
		}
	}
	return nil
}

func addPendingPolls(ctx context.Context, accountId string, pollMessageIds []string, messageById map[string]chatspg.Message, res *dashboardpg.MonthlySummaryResponse) error {
	pollsResult, err := chats.SelectPolls(ctx, chatspg.PollSelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "message_id", Operator: "IN", Value: pollMessageIds}}}}},
	})
	if err != nil {
		return err
	}

	var pollIds []string
	pollByMessageId := map[string]chatspg.Poll{}
	for _, p := range pollsResult.Polls {
		pollIds = append(pollIds, p.Id)
		pollByMessageId[p.MessageId] = p
	}
	if len(pollIds) == 0 {
		return nil
	}

	votesResult, err := chats.SelectPollVotes(ctx, chatspg.PollVoteSelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{
			{Key: "poll_id", Operator: "IN", Value: pollIds},
			{Key: "account_id", Operator: "=", Value: accountId},
		}}}},
	})
	if err != nil {
		return err
	}
	votedPollIds := map[string]bool{}
	for _, v := range votesResult.PollVotes {
		votedPollIds[v.PollId] = true
	}

	for _, messageId := range pollMessageIds {
		poll, ok := pollByMessageId[messageId]
		if !ok || votedPollIds[poll.Id] {
			continue
		}
		// A closed poll can no longer be answered, so it is not something "to do".
		if poll.ClosesAt != nil && poll.ClosesAt.Before(time.Now()) {
			continue
		}
		res.TodoConsents = append(res.TodoConsents, dashboardpg.MonthlySummaryTodoConsent{
			MessageId: messageId, ChannelId: messageById[messageId].ChannelId, Content: poll.Question, Kind: "poll",
		})
	}
	return nil
}

func addPendingConsents(ctx context.Context, accountId string, consentMessageIds []string, messageById map[string]chatspg.Message, res *dashboardpg.MonthlySummaryResponse) error {
	var confirmed []chatspg.MessageConfirmation
	if err := gormdb.GormDB.WithContext(ctx).
		Where("message_id IN ? AND account_id = ?", consentMessageIds, accountId).
		Find(&confirmed).Error; err != nil {
		return err
	}
	confirmedMessageIds := map[string]bool{}
	for _, c := range confirmed {
		confirmedMessageIds[c.MessageId] = true
	}

	for _, messageId := range consentMessageIds {
		if confirmedMessageIds[messageId] {
			continue
		}
		content := messageById[messageId].Content
		if len(content) > 80 {
			content = content[:80] + "…"
		}
		res.TodoConsents = append(res.TodoConsents, dashboardpg.MonthlySummaryTodoConsent{
			MessageId: messageId, ChannelId: messageById[messageId].ChannelId, Content: content, Kind: "consent",
		})
	}
	return nil
}

func childIdsForParent(ctx context.Context, accountId string) ([]string, error) {
	result, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "parent_id", Operator: "=", Value: accountId}}}}},
	})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(result.ParentChildren))
	for _, pc := range result.ParentChildren {
		ids = append(ids, pc.ChildId)
	}
	return ids, nil
}

// memberChannelIds has no existing internal-package equivalent (chats.SelectChannels lists
// Channel rows, not membership rows) so this queries ChannelMember directly, same as
// chats.ConsentStatus does for MessageConfirmation.
func memberChannelIds(ctx context.Context, accountId string) ([]string, error) {
	var members []chatspg.ChannelMember
	if err := gormdb.GormDB.WithContext(ctx).Where("account_id = ?", accountId).Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.ChannelId)
	}
	return ids, nil
}

func inRange(t *time.Time, start, end time.Time) bool {
	return t != nil && !t.Before(start) && t.Before(end)
}

// Outstanding returns everything a parent still has to act on, regardless of month: unpaid payments
// (overdue ones included), event sign-ups whose deadline hasn't passed, and polls/consents (of the
// last 120 days) not answered yet. It reuses the monthly digest's own queries with a wide window,
// so the Todo screen and the dashboard card can never disagree about what is outstanding.
func Outstanding(ctx context.Context, accountId string) (res dashboardpg.MonthlySummaryResponse, err error) {
	if !modules.Organizations {
		return res, nil
	}
	now := time.Now()
	childIds, err := childIdsForParent(ctx, accountId)
	if err != nil {
		return res, err
	}
	far := now.AddDate(20, 0, 0)
	if err = addPayments(ctx, accountId, now.AddDate(-20, 0, 0), far, &res); err != nil {
		return res, err
	}
	if err = addEvents(ctx, accountId, childIds, now, far, &res); err != nil {
		return res, err
	}
	if err = addPollsAndConsents(ctx, accountId, now.AddDate(0, 0, -120), far, &res); err != nil {
		return res, err
	}
	return res, nil
}
