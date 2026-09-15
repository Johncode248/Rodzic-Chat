package caregivers

import (
	"context"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	caregiverspg "github.com/Johncode248/Rodzic-Chat/project/common/caregivers"
)

// Delete removes the caller's own profile for a type. Recommendations
// pointing at it are left as-is (orphaned but harmless -- same
// deliberate non-cascading choice as chats.ChannelDeleteCondition).
func Delete(ctx context.Context, accountId string, caregiverType caregiverspg.CaregiverType) error {
	_, err := gormdb.Delete[caregiverspg.Caregiver](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "account_id", Operator: "=", Value: accountId},
				{Key: "type", Operator: "=", Value: int(caregiverType)},
			}},
		},
	})
	return err
}
