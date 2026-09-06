package mealsgt

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// parseDate/dateRangeFilters mirror absencesgt's identical helpers (see
// project/service/transport/absences/helpers.go) -- duplicated rather than
// shared since these two transport packages don't otherwise depend on each
// other and the codebase's convention (childInOrganization, parentOwnsChild,
// etc.) is to duplicate small per-domain helpers rather than add a shared
// package for a handful of lines.
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

// childIdsMatchingQuery resolves child ids whose first/last name matches q
// (case-insensitive substring), scoped to organizationId -- MealCancellation
// only stores child_id with no denormalized name, so searching by child
// name is a two-step lookup: resolve ids here, then the caller filters
// meals by "child_id IN (...)".
func childIdsMatchingQuery(ctx context.Context, organizationId, q string) ([]string, error) {
	like := "%" + q + "%"

	result, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{
					{Key: "first_name", Operator: "ILIKE", Value: like},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				}},
				{Filters: []databasepg.Filter{
					{Key: "last_name", Operator: "ILIKE", Value: like},
					{Key: "organization_id", Operator: "=", Value: organizationId},
				}},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(result.Children))
	for _, child := range result.Children {
		ids = append(ids, child.Id)
	}

	return ids, nil
}

// childInOrganization verifies that childId belongs to organizationId,
// mirroring absencesgt's helper of the same name (children have no direct
// link back, so this is the multi-tenancy check every meal-cancellation
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
// report/view a meal cancellation for that child.
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
// scope a parent's meal-cancellation listing to only their own children.
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
