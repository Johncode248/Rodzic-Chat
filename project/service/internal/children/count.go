package children

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
)

// CountForOrganization returns how many children currently belong to
// organizationId -- used both to enforce ChildrenLimit on creation (see
// Create's ValidateCreateMutation) and to stop a director from lowering
// their limit below what they're already using (see
// project/service/internal/organizations.InitiateUpdateMutation).
func CountForOrganization(ctx context.Context, organizationId string) (count int, err error) {
	ctx = logger.Enter(ctx, "children.CountForOrganization", organizationId)
	defer logger.Exit(ctx, "children.CountForOrganization", &count, &err)

	rows, err := gormdb.Select[childrenpg.Child](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "organization_id", Operator: "=", Value: organizationId}}},
		},
	})
	if err != nil {
		return 0, err
	}

	return len(rows), nil
}
