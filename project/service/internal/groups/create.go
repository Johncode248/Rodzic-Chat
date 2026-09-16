package groups

import (
	"context"
	"fmt"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	groupspg "github.com/Johncode248/Rodzic-Chat/project/common/groups"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
)

func Create(ctx context.Context, argument groupspg.CreateArgument) (res groupspg.CreateResponse, err error) {
	ctx = logger.Enter(ctx, "groups.Create", argument)
	defer logger.Exit(ctx, "groups.Create", &res, &err)

	var (
		usesMutation = &groupspg.CreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	if err = InitiateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	if err = ValidateCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := RollbackCreateMutation(ctx, usesMutation); rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()

	if err = FinalizeCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return groupspg.CreateResponse{
		Group: usesMutation.Transient.Current,
	}, nil
}

func InitiateCreateMutation(ctx context.Context, mutation *groupspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.CreateInitiator", mutation)
	defer logger.Exit(ctx, "groups.CreateInitiator", &err)

	future := mutation.Foundation.Populate()

	mutation.Transient.Future = future

	return nil
}

func ValidateCreateMutation(ctx context.Context, mutation *groupspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.ValidateCreateMutation", mutation)
	defer logger.Exit(ctx, "groups.ValidateCreateMutation", &err)

	return nil
}

func FinalizeCreateMutation(ctx context.Context, mutation *groupspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.FinalizeCreateMutation", mutation)
	defer logger.Exit(ctx, "groups.FinalizeCreateMutation", &err)

	err = gormdb.Create(ctx, mutation.Transient.Future)
	if err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	if err = createDefaultChannels(ctx, mutation.Transient.Current); err != nil {
		return err
	}

	return nil
}

// createDefaultChannels creates the two chats every group gets
// automatically: "rodzice między sobą" (ChannelTypeParents, no staff ever
// -- see chats.checkPostingPermission -- and never deletable, see
// chats.InitiateChannelDeleteMutation) and an Announcement channel staff
// use to post to that group's parents. Both are scoped to this specific
// group (GroupId), per product decision -- each group gets its own pair,
// not one shared pair for the whole organization.
func createDefaultChannels(ctx context.Context, group groupspg.Group) error {
	return routines.Select(
		func() error {
			_, err := chats.CreateChannel(ctx, chatspg.ChannelCreateArgument{
				OrganizationId: group.OrganizationId,
				GroupId:        group.Id,
				Type:           chatspg.ChannelTypeParents,
				Name:           fmt.Sprintf("Rodzice — %s", group.Name),
			})
			return err
		},
		func() error {
			_, err := chats.CreateChannel(ctx, chatspg.ChannelCreateArgument{
				OrganizationId: group.OrganizationId,
				GroupId:        group.Id,
				Type:           chatspg.ChannelTypeAnnouncement,
				Name:           fmt.Sprintf("Ogłoszenia — %s", group.Name),
			})
			return err
		},
	)
}

func RollbackCreateMutation(ctx context.Context, mutation *groupspg.CreateMutation) (err error) {
	ctx = logger.Enter(ctx, "groups.RollbackCreateMutation", mutation)
	defer logger.Exit(ctx, "groups.RollbackCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeDeleteMutation(ctx, &groupspg.DeleteMutation{
				Foundation: groupspg.DeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{
								Filters: []databasepg.Filter{
									{
										Key:      "id",
										Operator: "=",
										Value:    mutation.Transient.Current.Id,
									},
								},
							},
						},
					},
				},
			})
		},
		func() (err error) {
			// createDefaultChannels runs its two chats.CreateChannel calls
			// concurrently -- if one succeeded before the other failed
			// (triggering this rollback), that channel would otherwise be
			// orphaned: pointing at a GroupId that no longer exists, with
			// the caller believing group creation failed entirely. Delete
			// by group_id rather than by a specific channel id so this is
			// correct whether zero, one, or both channels were created.
			//
			// Deliberately bypasses chats.DeleteChannel: that function
			// refuses to delete a ChannelTypeParents channel (it's the
			// user-facing "cannot be deleted" default chat), which is the
			// right rule for the public delete endpoint but wrong here --
			// this whole group (and hence this channel) never successfully
			// finished being created in the first place.
			if mutation.Transient.Current.Id == "" {
				return nil
			}
			_, err = gormdb.Delete[chatspg.Channel](ctx, databasepg.Query{
				Groups: []databasepg.Group{
					{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: mutation.Transient.Current.Id}}},
				},
			})
			return err
		},
	)
}
