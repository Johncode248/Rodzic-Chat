package teacherinvites

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/logger"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/groups"
)

// AssignTeacherToGroup sets a newly-registered teacher as a group's
// homeroom teacher, when the invite that created them specified one.
func AssignTeacherToGroup(ctx context.Context, groupId, teacherId, organizationId string) (err error) {
	ctx = logger.Enter(ctx, "teacherinvites.AssignTeacherToGroup", groupId, teacherId)
	defer logger.Exit(ctx, "teacherinvites.AssignTeacherToGroup", &err)

	query := databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: groupId}}},
		},
	}
	query = query.AddOrganizationAuth(organizationId)

	_, err = groups.Update(ctx, groupspg.UpdateArgument{
		Query:     query,
		TeacherId: &teacherId,
	})

	return err
}
