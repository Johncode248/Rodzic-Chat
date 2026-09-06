package absencesgt

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// parseDate accepts "2006-01-02" or full RFC3339, returning ok=false (not
// an error) when the value is empty or unparseable -- callers just skip
// the filter in that case rather than 400ing on a slightly malformed date.
func parseDate(value string) (t time.Time, ok bool) {
	if value == "" {
		return t, false
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, true
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed, true
	}
	return t, false
}

// dateRangeFilters returns the filters for "does [Absence.DateFrom,
// Absence.DateTo] overlap [from, to]" -- the standard interval-overlap
// condition, so ?date_from=&date_to= finds absences active on/within that
// range rather than requiring an exact match.
func dateRangeFilters(from, to string) []databasepg.Filter {
	var filters []databasepg.Filter
	if t, ok := parseDate(to); ok {
		filters = append(filters, databasepg.Filter{Key: "date_from", Operator: "<=", Value: t})
	}
	if t, ok := parseDate(from); ok {
		filters = append(filters, databasepg.Filter{Key: "date_to", Operator: ">=", Value: t})
	}
	return filters
}

// childInOrganization verifies that childId belongs to organizationId,
// mirroring parentchildgt's helper of the same name (children have no
// direct link back, so this is the multi-tenancy check every absence
// write/read that touches a specific child must perform).
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

// parentOwnsChild reports whether accountId is linked to childId via a
// ParentChild row -- the real membership test for whether a parent may
// report/view an absence for that child.
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

// childIdsForParent returns every child id linked to accountId, used to
// scope a parent's absence listing to only their own children.
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
