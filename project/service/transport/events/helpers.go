package eventsgt

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

var errChildNotInOrganization = errors.New("child does not belong to the caller's organization")
var errNotChildsParent = errors.New("caller is not a registered parent of this child")
var errEventNotInOrganization = errors.New("event does not belong to the caller's organization")

func childInOrganization(ctx context.Context, childId, organizationId string) (bool, error) {
	if childId == "" || organizationId == "" {
		return false, nil
	}

	result, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: childId},
						{Key: "organization_id", Operator: "=", Value: organizationId},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(result.Children) > 0, nil
}

func parentOwnsChild(ctx context.Context, accountId, childId string) (bool, error) {
	if accountId == "" || childId == "" {
		return false, nil
	}

	result, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "parent_id", Operator: "=", Value: accountId},
						{Key: "child_id", Operator: "=", Value: childId},
					},
				},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(result.ParentChildren) > 0, nil
}

func childIdsForParent(ctx context.Context, accountId string) ([]string, error) {
	result, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "parent_id", Operator: "=", Value: accountId},
					},
				},
			},
		},
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

// eventInOrganization verifies eventId belongs to organizationId and returns
// the event, since EventRegistration has no organization_id column of its
// own -- every registration write/read scoped to a specific event must
// check this way.
func eventInOrganization(ctx context.Context, eventId, organizationId string) (eventspg.Event, bool, error) {
	result, err := events.SelectEvents(ctx, eventspg.EventSelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "id", Operator: "=", Value: eventId},
						{Key: "organization_id", Operator: "=", Value: organizationId},
					},
				},
			},
		},
	})
	if err != nil {
		return eventspg.Event{}, false, err
	}
	if len(result.Events) == 0 {
		return eventspg.Event{}, false, nil
	}

	return result.Events[0], true, nil
}

// eventIdsInOrganization returns every event id in organizationId, used to
// scope registration list queries that have no direct organization filter.
func eventIdsInOrganization(ctx context.Context, organizationId string) ([]string, error) {
	result, err := events.SelectEvents(ctx, eventspg.EventSelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "organization_id", Operator: "=", Value: organizationId},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.Events))
	for _, event := range result.Events {
		ids = append(ids, event.Id)
	}

	return ids, nil
}
