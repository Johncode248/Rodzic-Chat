package chatinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	chatspg "github.com/Johncode248/Rodzic-Chat/project/common/chats"
	"github.com/google/uuid"
)

var ErrInviteNotUsable = errors.New("this invite has expired or was already used")
var ErrCannotAcceptOwnInvite = errors.New("you cannot accept your own invite")

// Accept adds accepterAccountId as a ChannelMember of the inviter's chat
// channel -- the path for a co-parent who already has a Rodzic Chat
// account. Single-use: consuming it marks ConsumedByAccountId so the same
// token can't be redeemed twice.
func Accept(ctx context.Context, token string, accepterAccountId string) (res chatinvitespg.AcceptResponse, err error) {
	ctx = logger.Enter(ctx, "chatinvites.Accept", token, accepterAccountId)
	defer logger.Exit(ctx, "chatinvites.Accept", &res, &err)

	invite, err := getByToken(ctx, token)
	if err != nil {
		return res, err
	}
	if !invite.IsUsable(time.Now()) {
		return res, ErrInviteNotUsable
	}
	if invite.InviterAccountId == accepterAccountId {
		return res, ErrCannotAcceptOwnInvite
	}

	now := time.Now()
	member := chatspg.ChannelMember{
		Id:         uuid.NewString(),
		ChannelId:  invite.ChannelId,
		AccountId:  accepterAccountId,
		Role:       chatspg.ChannelMemberRoleMember,
		CreateTime: now,
	}
	if err = gormdb.Create(ctx, &member); err != nil {
		return res, err
	}

	invite.ConsumedByAccountId = accepterAccountId
	invite.ConsumedAt = &now
	if err = gormdb.Update(ctx, invite, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: invite.Id}}},
		},
	}); err != nil {
		return res, err
	}

	channelResult, err := gormdb.Select[chatspg.Channel](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: invite.ChannelId}}},
		},
	})
	if err != nil {
		return res, err
	}
	if len(channelResult) == 0 {
		return res, errors.New("channel not found")
	}

	return chatinvitespg.AcceptResponse{Channel: channelResult[0]}, nil
}
