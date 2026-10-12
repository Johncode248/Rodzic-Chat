package childevents

import (
	"context"
	"errors"
	"sort"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	childeventspg "github.com/Johncode248/Rodzic-Chat/project/common/childevents"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

var (
	ErrNoAttendees        = errors.New("choose at least one attendee: a child, yourself or another parent")
	ErrNotACoParent       = errors.New("that parent is not linked to any of your children")
	ErrEndBeforeStart     = errors.New("end_time must not be before start_time")
	ErrBadRecurrence      = errors.New("recurrence must be one of: daily, weekly, monthly, yearly (or empty)")
	ErrBadInterval        = errors.New("recurrence_interval must be between 1 and 99")
	ErrRecurrenceUntilBad = errors.New("recurrence_until must not be before start_time")
)

func inQuery(key string, values []string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: key, Operator: "IN", Value: values}}}}}
}

func eq(key, value string) databasepg.Query {
	return databasepg.Query{Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: key, Operator: "=", Value: value}}}}}
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// CoParentIds returns every OTHER parent linked to at least one of accountId's children
// ("drugi rodzic"), i.e. the only parents an event may be shared with.
func CoParentIds(ctx context.Context, accountId string) (map[string]bool, error) {
	mine, err := parentchild.Select(ctx, parentchildpg.SelectArgument{Query: eq("parent_id", accountId)})
	if err != nil {
		return nil, err
	}
	childIds := make([]string, 0, len(mine.ParentChildren))
	for _, link := range mine.ParentChildren {
		childIds = append(childIds, link.ChildId)
	}
	result := map[string]bool{}
	if len(childIds) == 0 {
		return result, nil
	}
	links, err := parentchild.Select(ctx, parentchildpg.SelectArgument{Query: inQuery("child_id", childIds)})
	if err != nil {
		return nil, err
	}
	for _, link := range links.ParentChildren {
		if link.ParentId != accountId {
			result[link.ParentId] = true
		}
	}
	return result, nil
}

// attendeesFrom validates and normalizes the attendee selection of a create/update: every child
// must be the caller's, every extra parent a co-parent of the caller.
func attendeesFrom(ctx context.Context, accountId string, childIds []string, includeMe bool, parentIds []string) (children, parents []string, err error) {
	children = unique(childIds)
	for _, childId := range children {
		isParent, checkErr := parentchild.IsParent(ctx, accountId, childId)
		if checkErr != nil {
			return nil, nil, checkErr
		}
		if !isParent {
			return nil, nil, ErrNotYourChild
		}
	}

	extra := unique(parentIds)
	if len(extra) > 0 {
		co, coErr := CoParentIds(ctx, accountId)
		if coErr != nil {
			return nil, nil, coErr
		}
		for _, id := range extra {
			if id != accountId && !co[id] {
				return nil, nil, ErrNotACoParent
			}
		}
	}

	parents = []string{}
	if includeMe {
		parents = append(parents, accountId)
	}
	for _, id := range extra {
		if id != accountId {
			parents = append(parents, id)
		}
	}

	if len(children) == 0 && len(parents) == 0 {
		return nil, nil, ErrNoAttendees
	}
	return children, parents, nil
}

func validateRecurrence(freq string, interval int) (string, int, error) {
	switch freq {
	case childeventspg.RecurrenceNone, childeventspg.RecurrenceDaily, childeventspg.RecurrenceWeekly,
		childeventspg.RecurrenceMonthly, childeventspg.RecurrenceYearly:
	default:
		return "", 0, ErrBadRecurrence
	}
	if interval == 0 {
		interval = 1
	}
	if interval < 1 || interval > 99 {
		return "", 0, ErrBadInterval
	}
	return freq, interval, nil
}

// replaceAttendees rewrites the two join tables of an event.
func replaceAttendees(ctx context.Context, eventId string, children, parents []string) error {
	if _, err := gormdb.Delete[childeventspg.ChildEventChild](ctx, eq("event_id", eventId)); err != nil {
		return err
	}
	if _, err := gormdb.Delete[childeventspg.ChildEventParent](ctx, eq("event_id", eventId)); err != nil {
		return err
	}
	for _, childId := range children {
		if err := gormdb.Create(ctx, &childeventspg.ChildEventChild{EventId: eventId, ChildId: childId}); err != nil {
			return err
		}
	}
	for _, accountId := range parents {
		if err := gormdb.Create(ctx, &childeventspg.ChildEventParent{EventId: eventId, AccountId: accountId}); err != nil {
			return err
		}
	}
	return nil
}

// LoadAttendees fills ChildIds/ParentIds on every event (2 queries total).
func LoadAttendees(ctx context.Context, events []childeventspg.ChildEvent) error {
	if len(events) == 0 {
		return nil
	}
	ids := make([]string, len(events))
	index := map[string]int{}
	for i, e := range events {
		ids[i] = e.Id
		index[e.Id] = i
		events[i].ChildIds = nil
		events[i].ParentIds = nil
	}
	childLinks, err := gormdb.Select[childeventspg.ChildEventChild](ctx, inQuery("event_id", ids))
	if err != nil {
		return err
	}
	for _, link := range childLinks {
		events[index[link.EventId]].ChildIds = append(events[index[link.EventId]].ChildIds, link.ChildId)
	}
	parentLinks, err := gormdb.Select[childeventspg.ChildEventParent](ctx, inQuery("event_id", ids))
	if err != nil {
		return err
	}
	for _, link := range parentLinks {
		events[index[link.EventId]].ParentIds = append(events[index[link.EventId]].ParentIds, link.AccountId)
	}
	for i := range events {
		sort.Strings(events[i].ChildIds)
		sort.Strings(events[i].ParentIds)
	}
	return nil
}

// canEdit: the creator, an attending parent, or a parent of any attending child.
func canEdit(ctx context.Context, accountId string, event childeventspg.ChildEvent) (bool, error) {
	if event.CreatedByAccountId == accountId {
		return true, nil
	}
	for _, id := range event.ParentIds {
		if id == accountId {
			return true, nil
		}
	}
	childIds := event.ChildIds
	if len(childIds) == 0 && event.ChildId != "" {
		childIds = []string{event.ChildId}
	}
	for _, childId := range childIds {
		isParent, err := parentchild.IsParent(ctx, accountId, childId)
		if err != nil {
			return false, err
		}
		if isParent {
			return true, nil
		}
	}
	return false, nil
}

func parseTime(value string) (time.Time, error) { return time.Parse(time.RFC3339, value) }
