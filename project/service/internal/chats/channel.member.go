package chats

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

var ErrNotChannelAdmin = errors.New("only the channel's admin may do this")
var ErrCannotRemoveAdmin = errors.New("the channel admin cannot be removed -- they can only leave")

// IsChannelMember reports whether accountId has a ChannelMember row for
// channelId -- the authorization check for the org-less parent-channel
// routes (see project/service/transport/chats/parentchannel.*.go), the
// counterpart of the OrganizationId check used everywhere else in this
// package.
func IsChannelMember(ctx context.Context, channelId, accountId string) (bool, error) {
	member, err := getMember(ctx, channelId, accountId)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}

func getMember(ctx context.Context, channelId, accountId string) (*chatspg.ChannelMember, error) {
	members, err := gormdb.Select[chatspg.ChannelMember](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "channel_id", Operator: "=", Value: channelId},
				{Key: "account_id", Operator: "=", Value: accountId},
			}},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, nil
	}
	return &members[0], nil
}

// IsChannelAdmin reports whether accountId is the admin (creator) of channelId.
func IsChannelAdmin(ctx context.Context, channelId, accountId string) (bool, error) {
	member, err := getMember(ctx, channelId, accountId)
	if err != nil {
		return false, err
	}
	return member != nil && member.Role == chatspg.ChannelMemberRoleAdmin, nil
}

// CountAdminChannels reports how many org-less parent channels accountId
// has created -- the "3 kanały" cap is enforced against this in
// chatinvites.Create.
func CountAdminChannels(ctx context.Context, accountId string) (int, error) {
	members, err := gormdb.Select[chatspg.ChannelMember](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "account_id", Operator: "=", Value: accountId},
				{Key: "role", Operator: "=", Value: int(chatspg.ChannelMemberRoleAdmin)},
			}},
		},
	})
	if err != nil {
		return 0, err
	}
	return len(members), nil
}

// ListMembers returns a channel's roster with display names attached, newest first isn't
// meaningful here so it's returned admin-first then by join order.
func ListMembers(ctx context.Context, channelId string) (res chatspg.ChannelMemberListResponse, err error) {
	members, err := gormdb.Select[chatspg.ChannelMember](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "channel_id", Operator: "=", Value: channelId}}},
		},
		OrderBy: databasepg.Order{Key: "create_time", Direction: databasepg.DirectionAscending},
	})
	if err != nil {
		return res, err
	}

	accountIds := make([]string, 0, len(members))
	for _, m := range members {
		accountIds = append(accountIds, m.AccountId)
	}
	nameById := map[string]string{}
	if len(accountIds) > 0 {
		accountResult, accErr := accounts.Select(ctx, accountspg.SelectArgument{
			Query: databasepg.Query{
				Groups: []databasepg.Group{{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: accountIds}}}},
			},
		})
		if accErr != nil {
			return res, accErr
		}
		for _, a := range accountResult.Accounts {
			nameById[a.Id] = a.Name
		}
	}

	infos := make([]chatspg.ChannelMemberInfo, 0, len(members))
	for _, m := range members {
		// Admin first, then join order -- a small, deliberate re-sort after the
		// create_time-ordered query above (admin is nearly always first anyway
		// since they created the channel, but this makes it a guarantee).
		infos = append(infos, chatspg.ChannelMemberInfo{AccountId: m.AccountId, DisplayName: nameById[m.AccountId], Role: m.Role})
	}
	for i := 1; i < len(infos); i++ {
		if infos[i].Role == chatspg.ChannelMemberRoleAdmin {
			admin := infos[i]
			copy(infos[1:i+1], infos[0:i])
			infos[0] = admin
			break
		}
	}

	return chatspg.ChannelMemberListResponse{Members: infos}, nil
}

// RemoveMember handles both "leave" (targetAccountId == callerAccountId, always allowed for a
// non-admin; an admin leaving is also allowed -- the channel simply keeps going with no admin,
// see this package's doc notes) and "kick" (targetAccountId != callerAccountId, which requires
// callerAccountId to be the channel's admin, and can never target the admin themselves).
func RemoveMember(ctx context.Context, channelId, targetAccountId, callerAccountId string) error {
	target, err := getMember(ctx, channelId, targetAccountId)
	if err != nil {
		return err
	}
	if target == nil {
		return nil // already not a member -- idempotent
	}

	if targetAccountId != callerAccountId {
		callerIsAdmin, adminErr := IsChannelAdmin(ctx, channelId, callerAccountId)
		if adminErr != nil {
			return adminErr
		}
		if !callerIsAdmin {
			return ErrNotChannelAdmin
		}
		if target.Role == chatspg.ChannelMemberRoleAdmin {
			return ErrCannotRemoveAdmin
		}
	}

	_, err = gormdb.Delete[chatspg.ChannelMember](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "channel_id", Operator: "=", Value: channelId},
				{Key: "account_id", Operator: "=", Value: targetAccountId},
			}},
		},
	})
	return err
}
