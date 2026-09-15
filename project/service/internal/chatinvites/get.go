package chatinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

var ErrInviteNotFound = errors.New("invite not found")

func byToken(token string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "token", Operator: "=", Value: token}}},
	}}
}

func getByToken(ctx context.Context, token string) (invite chatinvitespg.ChatInvite, err error) {
	existing, err := gormdb.Select[chatinvitespg.ChatInvite](ctx, byToken(token))
	if err != nil {
		return invite, err
	}
	if len(existing) == 0 {
		return invite, ErrInviteNotFound
	}
	return existing[0], nil
}

// GetPublicInfo never errors on an invalid/expired/used token -- see
// chatinvitespg.PublicInfoResponse's doc comment.
func GetPublicInfo(ctx context.Context, token string) (res chatinvitespg.PublicInfoResponse, err error) {
	ctx = logger.Enter(ctx, "chatinvites.GetPublicInfo", token)
	defer logger.Exit(ctx, "chatinvites.GetPublicInfo", &res, &err)

	invite, err := getByToken(ctx, token)
	if err != nil {
		if err == ErrInviteNotFound {
			return chatinvitespg.PublicInfoResponse{Usable: false, Reason: "Nieprawidłowy link zaproszenia"}, nil
		}
		return res, err
	}

	if !invite.IsUsable(time.Now()) {
		reason := "To zaproszenie wygasło"
		if invite.ConsumedByAccountId != "" {
			reason = "To zaproszenie zostało już wykorzystane"
		}
		return chatinvitespg.PublicInfoResponse{Usable: false, Reason: reason}, nil
	}

	inviterResult, err := accounts.Select(ctx, accountspg.SelectArgument{
		Query: databasepg.Query{Limit: 1, Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: invite.InviterAccountId}}},
		}},
	})
	if err != nil {
		return res, err
	}

	inviterName := ""
	if len(inviterResult.Accounts) > 0 {
		inviterName = inviterResult.Accounts[0].Name
	}

	return chatinvitespg.PublicInfoResponse{InviterName: inviterName, Usable: true}, nil
}
