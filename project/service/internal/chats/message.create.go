package chats

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/routines"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	childrenpg "github.com/Johncode248/Rodzic-Chat/project/common/children"
	parentchildpg "github.com/Johncode248/Rodzic-Chat/project/common/parentchild"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/children"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/parentchild"
)

// rateLimitWindow/rateLimitMax are the anti-spam guardrail: at most 5
// messages per sender per channel per 60 seconds. This threshold is a
// judgment call and deliberately not wired up to any config system yet --
// tune it here if it turns out too strict/loose in practice.
const rateLimitWindow = 60 * time.Second
const rateLimitMax = 5

var ErrRateLimited = errors.New("Zbyt wiele wiadomości w krótkim czasie, spróbuj ponownie za chwilę")
var ErrAnnouncementStaffOnly = errors.New("only staff may post new messages to an announcement channel")
var ErrNotGroupMember = errors.New("caller is not a member of this group channel")
var ErrParentsChannelStaffExcluded = errors.New("staff may not post in the default parents-only channel")
var ErrChannelNotInOrganization = errors.New("channel does not belong to the caller's organization")

// CreateMessage is the "smart, not spammy" chat entry point: it resolves
// the target channel, enforces who may post into it (staff-only for
// announcement channels; group-membership for group channels), applies the
// rate limit, and only then creates the Message row.
func CreateMessage(ctx context.Context, argument chatspg.MessageCreateArgument, accountType int) (res chatspg.MessageCreateResponse, err error) {
	ctx = logger.Enter(ctx, "chats.CreateMessage", argument)
	defer logger.Exit(ctx, "chats.CreateMessage", &res, &err)

	channelResult, err := SelectChannels(ctx, chatspg.ChannelSelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: argument.ChannelId}}},
			},
		},
	})
	if err != nil {
		return res, err
	}
	if len(channelResult.Channels) == 0 || channelResult.Channels[0].OrganizationId != argument.OrganizationId {
		return res, ErrChannelNotInOrganization
	}

	channel := channelResult.Channels[0]

	if err = checkPostingPermission(ctx, channel, argument.SenderId, accountType); err != nil {
		return res, err
	}

	if err = checkRateLimit(ctx, argument.ChannelId, argument.SenderId); err != nil {
		return res, err
	}

	if argument.ReplyToId != "" {
		target, loadErr := LoadMessage(ctx, argument.ReplyToId)
		if loadErr != nil || target.ChannelId != argument.ChannelId {
			return res, errors.New("reply target must be a message in the same channel")
		}
	}

	var (
		usesMutation = &chatspg.MessageCreateMutation{
			Foundation: argument,
		}
	)

	if err = argument.Validate(); err != nil {
		return res, err
	}

	usesMutation.Transient.Future = usesMutation.Foundation.Populate()

	if err = FinalizeMessageCreateMutation(ctx, usesMutation); err != nil {
		return res, err
	}

	return chatspg.MessageCreateResponse{
		Message: usesMutation.Transient.Current,
	}, nil
}

// checkPostingPermission implements the per-channel-type posting rule:
// Announcement channels only take new messages from staff (Director/
// Teacher); Group channels take messages from any staff member, or from a
// parent who has at least one child actually in that group; Direct
// channels are unrestricted here (a full "who can DM whom" access-list
// system is out of scope for this version).
func checkPostingPermission(ctx context.Context, channel chatspg.Channel, senderId string, accountType int) error {
	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()

	switch channel.Type {
	case chatspg.ChannelTypeAnnouncement:
		if !isStaff {
			return ErrAnnouncementStaffOnly
		}
		return nil
	case chatspg.ChannelTypeGroup:
		if isStaff {
			return nil
		}
		if channel.GroupId == "" {
			// organization-wide group channel with no specific GroupId:
			// treat like an announcement-less general channel, open to
			// any org member.
			return nil
		}

		return requireParentInGroup(ctx, channel.GroupId, senderId)
	case chatspg.ChannelTypeParents:
		// The default "rodzice między sobą" channel: staff (any type) are
		// never allowed to post here, regardless of group membership --
		// unlike ChannelTypeGroup, which lets staff post unconditionally.
		if isStaff {
			return ErrParentsChannelStaffExcluded
		}
		return requireParentInGroup(ctx, channel.GroupId, senderId)
	default:
		return nil
	}
}

// requireParentInGroup checks that senderId is linked (as a parent) to at
// least one child in groupId -- the membership test shared by
// ChannelTypeGroup and ChannelTypeParents.
func requireParentInGroup(ctx context.Context, groupId, senderId string) error {
	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "group_id", Operator: "=", Value: groupId}}},
			},
		},
	})
	if err != nil {
		return err
	}

	childIds := make([]string, 0, len(childrenResult.Children))
	for _, child := range childrenResult.Children {
		childIds = append(childIds, child.Id)
	}

	if len(childIds) == 0 {
		return ErrNotGroupMember
	}

	linkResult, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Limit: 1,
			Groups: []databasepg.Group{
				{
					Filters: []databasepg.Filter{
						{Key: "parent_id", Operator: "=", Value: senderId},
						{Key: "child_id", Operator: "IN", Value: childIds},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if len(linkResult.ParentChildren) == 0 {
		return ErrNotGroupMember
	}

	return nil
}

// CheckReadPermission implements the per-channel-type read rule, mirroring
// checkPostingPermission but for reads: staff (any type) may read every
// channel in their organization. A parent may freely read Announcement/
// Direct channels and organization-wide Group channels (GroupId == ""), but
// a Group/Parents channel tied to a specific GroupId requires the parent to
// actually have a child in that group -- without this, any parent could
// read another class's messages, reactions, read-receipts and polls just by
// knowing/guessing that channel's id.
func CheckReadPermission(ctx context.Context, channel chatspg.Channel, accountId string, accountType int) error {
	isStaff := accountType == accountspg.AccountTypeDirector.Ordinal() || accountType == accountspg.AccountTypeCoDirector.Ordinal() || accountType == accountspg.AccountTypeTeacher.Ordinal()
	if isStaff {
		return nil
	}

	switch channel.Type {
	case chatspg.ChannelTypeGroup, chatspg.ChannelTypeParents:
		if channel.GroupId == "" {
			return nil
		}
		return requireParentInGroup(ctx, channel.GroupId, accountId)
	default:
		return nil
	}
}

// ParentGroupIds returns the distinct group ids of every child linked to
// accountId as a parent -- used to bulk-filter a channel list down to the
// groups that parent may actually see, without one requireParentInGroup
// query per channel.
func ParentGroupIds(ctx context.Context, accountId string) ([]string, error) {
	linkResult, err := parentchild.Select(ctx, parentchildpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "parent_id", Operator: "=", Value: accountId}}},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	childIds := make([]string, 0, len(linkResult.ParentChildren))
	for _, pc := range linkResult.ParentChildren {
		childIds = append(childIds, pc.ChildId)
	}
	if len(childIds) == 0 {
		return nil, nil
	}

	childrenResult, err := children.Select(ctx, childrenpg.SelectArgument{
		Query: databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: childIds}}},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	groupIds := make([]string, 0, len(childrenResult.Children))
	for _, child := range childrenResult.Children {
		if child.GroupId == "" || seen[child.GroupId] {
			continue
		}
		seen[child.GroupId] = true
		groupIds = append(groupIds, child.GroupId)
	}

	return groupIds, nil
}

func checkRateLimit(ctx context.Context, channelId, senderId string) error {
	since := time.Now().Add(-rateLimitWindow)

	recent, err := gormdb.Select[chatspg.Message](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{
				Filters: []databasepg.Filter{
					{Key: "channel_id", Operator: "=", Value: channelId},
					{Key: "sender_id", Operator: "=", Value: senderId},
					{Key: "create_time", Operator: ">=", Value: since},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	if len(recent) >= rateLimitMax {
		return ErrRateLimited
	}

	return nil
}

func FinalizeMessageCreateMutation(ctx context.Context, mutation *chatspg.MessageCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.FinalizeMessageCreateMutation", mutation)
	defer logger.Exit(ctx, "chats.FinalizeMessageCreateMutation", &err)

	if err = gormdb.Create(ctx, mutation.Transient.Future); err != nil {
		return err
	}

	mutation.Transient = mutation.Transient.Update()

	return nil
}

// RollbackMessageCreateMutation is used by poll creation to undo the
// message row if the subsequent Poll/PollOption creation fails.
func RollbackMessageCreateMutation(ctx context.Context, mutation *chatspg.MessageCreateMutation) (err error) {
	ctx = logger.Enter(ctx, "chats.RollbackMessageCreateMutation", mutation)
	defer logger.Exit(ctx, "chats.RollbackMessageCreateMutation", &err)

	return routines.Select(
		func() (err error) {
			return FinalizeMessageDeleteMutation(ctx, &chatspg.MessageDeleteMutation{
				Foundation: chatspg.MessageDeleteArgument{
					Query: databasepg.Query{
						Groups: []databasepg.Group{
							{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: mutation.Transient.Current.Id}}},
						},
					},
				},
			})
		},
	)
}
