package calendar

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	calendarpg "github.com/Johncode248/Rodzic-Chat/project/common/calendar"
	eventspg "github.com/Johncode248/Rodzic-Chat/project/common/events"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/childevents"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
)

// Get builds the parent-facing calendar feed: every personal ChildEvent for
// every child the caller parents, PLUS every placowka Event belonging to
// an organization/group one of those children is actually in -- merged
// into one chronological-ish list (see calendarpg.Item). Read-only on the
// placowka side; personal events are the only ones a parent can edit here.
func Get(ctx context.Context, accountId string) (res calendarpg.GetResponse, err error) {
	ctx = logger.Enter(ctx, "calendar.Get", accountId)
	defer logger.Exit(ctx, "calendar.Get", &res, &err)

	mine, err := children.SelectMine(ctx, accountId)
	if err != nil {
		return res, err
	}

	childNameById := map[string]string{}
	childIds := make([]string, 0, len(mine.Children))
	// organizationId -> groupIds (including "" for org-wide) that at least
	// one of the caller's children actually belongs to, keyed per child so
	// a placowka event can be attributed back to the right child(ren).
	type orgMembership struct {
		childId, organizationId, groupId string
	}
	var memberships []orgMembership

	for _, child := range mine.Children {
		childNameById[child.Id] = child.FirstName + " " + child.LastName
		childIds = append(childIds, child.Id)
		if child.OrganizationId != "" {
			memberships = append(memberships, orgMembership{
				childId: child.Id, organizationId: child.OrganizationId, groupId: child.GroupId,
			})
		}
	}

	var items []calendarpg.Item

	personalEvents, err := childevents.ListForChildren(ctx, childIds)
	if err != nil {
		return res, err
	}
	for _, event := range personalEvents {
		items = append(items, calendarpg.Item{
			Id:          event.Id,
			Source:      calendarpg.SourcePersonal,
			ChildId:     event.ChildId,
			ChildName:   childNameById[event.ChildId],
			Title:       event.Title,
			Description: event.Description,
			Location:    event.Location,
			StartTime:   event.StartTime,
			EndTime:     event.EndTime,
		})
	}

	// One query per distinct organization the caller's children belong to
	// -- in practice 0 or 1 for the overwhelming majority of parents.
	organizationIds := map[string]bool{}
	for _, m := range memberships {
		organizationIds[m.organizationId] = true
	}
	for organizationId := range organizationIds {
		orgEvents, selectErr := gormdb.Select[eventspg.Event](ctx, databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "organization_id", Operator: "=", Value: organizationId}}},
			},
		})
		if selectErr != nil {
			return res, selectErr
		}
		for _, event := range orgEvents {
			for _, m := range memberships {
				if m.organizationId != organizationId {
					continue
				}
				// Org-wide event (no GroupId) applies to every child in
				// that org; a group-scoped event only applies to children
				// actually in that group.
				if event.GroupId != "" && event.GroupId != m.groupId {
					continue
				}
				items = append(items, calendarpg.Item{
					Id:          event.Id,
					Source:      calendarpg.SourcePlacowka,
					ChildId:     m.childId,
					ChildName:   childNameById[m.childId],
					Title:       event.Title,
					Description: event.Description,
					StartTime:   event.EventDate,
				})
			}
		}
	}

	return calendarpg.GetResponse{Items: items}, nil
}
