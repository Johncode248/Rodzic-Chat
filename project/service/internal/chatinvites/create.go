package chatinvites

import (
	"context"
	"errors"
	"fmt"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/chats"
	"github.com/google/uuid"
)

const expiryDays = 7

// MaxChannelsPerAccount caps how many org-less parent channels one account
// may create (be the admin of) -- a parent can still join as many other
// people's channels as they like, this only limits how many they start
// themselves. A product judgment call to keep this feature from becoming
// an unbounded free group-chat host; revisit if it turns out too strict.
const MaxChannelsPerAccount = 3

var ErrChannelLimitReached = errors.New("możesz utworzyć maksymalnie 3 kanały -- opuść jeden, aby założyć nowy")

func resolveDisplayName(ctx context.Context, accountId string) string {
	if accountResult, accErr := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{Limit: 1, Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: accountId}}},
		}},
	}); accErr == nil && len(accountResult.Accounts) > 0 && accountResult.Accounts[0].Name != "" {
		return accountResult.Accounts[0].Name
	}
	return "Rodzic"
}

// Create sets up a brand-new, org-less channel for accountId (Type Direct,
// OrganizationId "") -- the "rodzice między sobą" case where parents share
// no placówka -- adds accountId as its Admin ChannelMember, and issues a
// token to share with the first invitee (see CreateForChannel to invite
// more people into an already-existing channel instead of starting a new
// one). Capped at [MaxChannelsPerAccount] channels created per account.
func Create(ctx context.Context, accountId string) (invite chatinvitespg.ChatInvite, channel chatspg.Channel, err error) {
	ctx = logger.Enter(ctx, "chatinvites.Create", accountId)
	defer logger.Exit(ctx, "chatinvites.Create", &invite, &channel, &err)

	adminCount, err := chats.CountAdminChannels(ctx, accountId)
	if err != nil {
		return invite, channel, err
	}
	if adminCount >= MaxChannelsPerAccount {
		return invite, channel, ErrChannelLimitReached
	}

	now := time.Now()
	inviterName := resolveDisplayName(ctx, accountId)

	channel = chatspg.Channel{
		Id:         uuid.NewString(),
		Type:       chatspg.ChannelTypeDirect,
		Name:       fmt.Sprintf("Czat: %s", inviterName),
		CreateTime: now,
		UpdateTime: now,
	}
	if err = gormdb.Create(ctx, &channel); err != nil {
		return invite, channel, err
	}

	member := chatspg.ChannelMember{
		Id:         uuid.NewString(),
		ChannelId:  channel.Id,
		AccountId:  accountId,
		Role:       chatspg.ChannelMemberRoleAdmin,
		CreateTime: now,
	}
	if err = gormdb.Create(ctx, &member); err != nil {
		return invite, channel, err
	}

	invite = chatinvitespg.ChatInvite{
		Id:               uuid.NewString(),
		Token:            random.GenerateToken(),
		ChannelId:        channel.Id,
		InviterAccountId: accountId,
		ExpiresAt:        now.AddDate(0, 0, expiryDays),
		CreateTime:       now,
	}
	if err = gormdb.Create(ctx, &invite); err != nil {
		return invite, channel, err
	}

	return invite, channel, nil
}

// CreateForChannel issues an additional invite token for an existing
// channel -- "Zaproś kolejną osobę" from within the channel's settings,
// once it already has at least one member beyond its creator. Only the
// channel's admin may do this.
func CreateForChannel(ctx context.Context, channelId, accountId string) (invite chatinvitespg.ChatInvite, err error) {
	ctx = logger.Enter(ctx, "chatinvites.CreateForChannel", channelId, accountId)
	defer logger.Exit(ctx, "chatinvites.CreateForChannel", &invite, &err)

	isAdmin, err := chats.IsChannelAdmin(ctx, channelId, accountId)
	if err != nil {
		return invite, err
	}
	if !isAdmin {
		return invite, chats.ErrNotChannelAdmin
	}

	now := time.Now()
	invite = chatinvitespg.ChatInvite{
		Id:               uuid.NewString(),
		Token:            random.GenerateToken(),
		ChannelId:        channelId,
		InviterAccountId: accountId,
		ExpiresAt:        now.AddDate(0, 0, expiryDays),
		CreateTime:       now,
	}
	if err = gormdb.Create(ctx, &invite); err != nil {
		return invite, err
	}
	return invite, nil
}
