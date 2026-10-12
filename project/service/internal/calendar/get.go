package calendar

import (
	"context"
	"sort"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
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

	personalEvents, err := childevents.ListForAccount(ctx, accountId, childIds)
	if err != nil {
		return res, err
	}

	// Names of every parent involved (the caller, co-parents, attending parents).
	coParents, err := childevents.CoParentIds(ctx, accountId)
	if err != nil {
		return res, err
	}
	personIds := map[string]bool{accountId: true}
	for id := range coParents {
		personIds[id] = true
	}
	for _, event := range personalEvents {
		for _, id := range event.ParentIds {
			personIds[id] = true
		}
	}
	idList := make([]string, 0, len(personIds))
	for id := range personIds {
		idList = append(idList, id)
	}
	accountRows, err := gormdb.Select[accountspg.Account](ctx, databasepg.Query{
		Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: idList}}}},
	})
	if err != nil {
		return res, err
	}
	personName := map[string]string{}
	for _, row := range accountRows {
		personName[row.Id] = row.Name
	}

	windowFrom := time.Now().AddDate(0, 0, -60)
	windowTo := time.Now().AddDate(0, 0, 400)
	for _, event := range personalEvents {
		childIdsOfEvent := event.ChildIds
		if len(childIdsOfEvent) == 0 && event.ChildId != "" {
			childIdsOfEvent = []string{event.ChildId}
		}
		var childNames []string
		for _, id := range childIdsOfEvent {
			if name, ok := childNameById[id]; ok {
				childNames = append(childNames, name)
			}
		}
		var parentNames []string
		for _, id := range event.ParentIds {
			parentNames = append(parentNames, personName[id])
		}
		firstChild, firstChildName := "", ""
		if len(childIdsOfEvent) > 0 {
			firstChild = childIdsOfEvent[0]
			firstChildName = childNameById[firstChild]
		}

		for occurrence, span := range childevents.Occurrences(event, windowFrom, windowTo) {
			items = append(items, calendarpg.Item{
				Id:                 event.Id,
				Source:             calendarpg.SourcePersonal,
				ChildId:            firstChild,
				ChildName:          firstChildName,
				Title:              event.Title,
				Description:        event.Description,
				Location:           event.Location,
				StartTime:          span.Start,
				EndTime:            span.End,
				ChildIds:           childIdsOfEvent,
				ChildNames:         childNames,
				ParentIds:          event.ParentIds,
				ParentNames:        parentNames,
				AllDay:             event.AllDay,
				Recurrence:         event.Recurrence,
				RecurrenceInterval: event.RecurrenceInterval,
				RecurrenceUntil:    event.RecurrenceUntil,
				Occurrence:         occurrence,
				EventStartTime:     event.StartTime,
				EventEndTime:       event.EndTime,
			})
		}
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

	response := calendarpg.GetResponse{Items: items, Me: calendarpg.Person{Id: accountId, Name: personName[accountId]}}
	for id := range coParents {
		response.CoParents = append(response.CoParents, calendarpg.Person{Id: id, Name: personName[id]})
	}
	sort.Slice(response.CoParents, func(i, j int) bool { return response.CoParents[i].Name < response.CoParents[j].Name })
	return response, nil
}
