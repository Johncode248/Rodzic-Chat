package children

import (
	"context"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

// UnenrollForOrganization clears OrganizationId/GroupId on every child
// currently belonging to organizationId -- called right before an
// Organization row is deleted (see
// project/service/internal/organizations.FinalizeDeleteMutation) so a
// deleted placówka never takes a parent's children away with it; the Child
// rows (and every ParentChild link to them) survive untouched.
func UnenrollForOrganization(ctx context.Context, organizationId string) (err error) {
	ctx = logger.Enter(ctx, "children.UnenrollForOrganization", organizationId)
	defer logger.Exit(ctx, "children.UnenrollForOrganization", &err)

	rows, err := gormdb.Select[childrenpg.Child](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "organization_id", Operator: "=", Value: organizationId}}},
		},
	})
	if err != nil {
		return err
	}

	now := time.Now()
	for _, child := range rows {
		child.OrganizationId = ""
		child.GroupId = ""
		child.UpdateTime = now
		if err = gormdb.Update(ctx, &child, byIdQuery(child.Id)); err != nil {
			return err
		}
	}

	return nil
}

// UnenrollForGroup clears GroupId (only) on every child currently assigned
// to groupId -- called right before a Group row is deleted (see
// project/service/internal/groups.FinalizeDeleteMutation). OrganizationId
// is left untouched: the child is still enrolled in the placówka, just no
// longer assigned to that (now-deleted) group.
func UnenrollForGroup(ctx context.Context, groupId string) (err error) {
	ctx = logger.Enter(ctx, "children.UnenrollForGroup", groupId)
	defer logger.Exit(ctx, "children.UnenrollForGroup", &err)

	rows, err := gormdb.Select[childrenpg.Child](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: groupId}}},
		},
	})
	if err != nil {
		return err
	}

	now := time.Now()
	for _, child := range rows {
		child.GroupId = ""
		child.UpdateTime = now
		if err = gormdb.Update(ctx, &child, byIdQuery(child.Id)); err != nil {
			return err
		}
	}

	return nil
}
