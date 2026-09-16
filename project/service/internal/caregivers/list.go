package caregivers

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
)

// List browses caregivers of one type, newest first, optionally narrowed
// by a free-text match against the display name or city. limit <= 0 means
// "no limit" (every existing caller before pagination was added), offset
// <= 0 means "from the start".
func List(ctx context.Context, caregiverType caregiverspg.CaregiverType, query string, limit int, offset int) (res caregiverspg.ListResponse, err error) {
	typeFilter := databasepg.Filter{Key: "type", Operator: "=", Value: int(caregiverType)}

	var groups []databasepg.Group
	if query == "" {
		groups = []databasepg.Group{{Filters: []databasepg.Filter{typeFilter}}}
	} else {
		// type=X AND (display_name ILIKE %q% OR city ILIKE %q%) -- expressed as two
		// OR'd groups, each ANDing the type filter back in (see databasepg.Query's
		// OR-of-ANDs shape in database/gormdb/condition.go).
		like := "%" + query + "%"
		groups = []databasepg.Group{
			{Filters: []databasepg.Filter{typeFilter, {Key: "display_name", Operator: "ILIKE", Value: like}}},
			{Filters: []databasepg.Filter{typeFilter, {Key: "city", Operator: "ILIKE", Value: like}}},
		}
	}

	rows, err := gormdb.Select[caregiverspg.Caregiver](ctx, databasepg.Query{
		Groups:  groups,
		OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionDescending},
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return res, err
	}

	summaries := make([]caregiverspg.Summary, 0, len(rows))
	for _, row := range rows {
		count, countErr := recommendationCount(ctx, row.Id)
		if countErr != nil {
			return res, countErr
		}
		summaries = append(summaries, caregiverspg.Summary{Caregiver: row, RecommendationCount: count})
	}

	return caregiverspg.ListResponse{Caregivers: summaries}, nil
}
