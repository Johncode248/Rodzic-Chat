package chats

import (
	"context"
	"errors"
	"sort"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

var ErrNotChannelAdmin = errors.New("only the channel's admin may do this")
var ErrCannotRemoveAdmin = errors.New("the channel admin cannot be removed -- they can only leave")
var ErrNoPermission = errors.New("you do not have permission to do this in this chat")
var ErrCannotModerateTarget = errors.New("you cannot do this to that member")

// IsChannelMember reports whether accountId has a ChannelMember row for
// channelId -- the authorization check for the org-less parent-channel
// routes (see project/service/transport/chats/parentchannel.*.go), the
// counterpart of the OrganizationId check used everywhere else in this
// package.
func IsChannelMember(ctx context.Context, channelId, accountId string) (bool, error) {
	member, err := getMember(ctx, channelId, accountId)
	if err != nil || member == nil {
		return false, err
	}
	// An active ban keeps the membership row but shuts the person out (see ChannelBan).
	ban, err := ActiveBan(ctx, channelId, accountId)
	if err != nil {
		return false, err
	}
	return ban == nil, nil
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

// ListMembers returns a channel's roster with display names attached, admin first then by join
// order, plus the caller's own role/permissions. Ban details are included only for callers who may
// see them (PermBan).
func ListMembers(ctx context.Context, channelId, callerAccountId string) (res chatspg.ChannelMemberListResponse, err error) {
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

	var caller *chatspg.ChannelMember
	for i := range members {
		if members[i].AccountId == callerAccountId {
			caller = &members[i]
		}
	}
	callerPerms := chatspg.ChannelPermission(0)
	if caller != nil {
		res.MyRole = caller.Role
		callerPerms = caller.EffectivePermissions()
		res.MyPermissions = callerPerms
	}
	bansById := map[string]chatspg.ChannelBan{}
	if callerPerms.Has(chatspg.PermBan) {
		bans, banErr := ListActiveBans(ctx, channelId)
		if banErr != nil {
			return res, banErr
		}
		for _, b := range bans {
			bansById[b.AccountId] = b
		}
	}

	infos := make([]chatspg.ChannelMemberInfo, 0, len(members))
	for _, m := range members {
		info := chatspg.ChannelMemberInfo{
			AccountId: m.AccountId, DisplayName: nameById[m.AccountId], Role: m.Role,
			Permissions: m.EffectivePermissions(), GrantedPermissions: m.Permissions,
		}
		if ban, ok := bansById[m.AccountId]; ok {
			info.BannedUntil = ban.Until
			info.BannedForever = ban.Until == nil
		}
		infos = append(infos, info)
	}
	// Admin first, then join order.
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].Role == chatspg.ChannelMemberRoleAdmin && infos[j].Role != chatspg.ChannelMemberRoleAdmin
	})
	res.Members = infos
	return res, nil
}

// RemoveMember handles both "leave" (targetAccountId == callerAccountId, always allowed; an admin
// leaving is also allowed -- the channel simply keeps going with no admin) and "kick"
// (targetAccountId != callerAccountId), which requires PermRemoveMembers and that the caller
// outranks the target (see canModerate).
func RemoveMember(ctx context.Context, channelId, targetAccountId, callerAccountId string) error {
	target, err := getMember(ctx, channelId, targetAccountId)
	if err != nil {
		return err
	}
	if target == nil {
		return nil // already not a member -- idempotent
	}

	if targetAccountId != callerAccountId {
		if err = requireModeration(ctx, channelId, callerAccountId, target, chatspg.PermRemoveMembers); err != nil {
			return err
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
