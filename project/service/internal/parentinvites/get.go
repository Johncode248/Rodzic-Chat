package parentinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

var ErrInviteNotFound = errors.New("invite not found")

func byToken(token string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "token", Operator: "=", Value: token}}},
	}}
}

func getByToken(ctx context.Context, token string) (invite parentinvitespg.ParentInvite, err error) {
	existing, err := gormdb.Select[parentinvitespg.ParentInvite](ctx, byToken(token))
	if err != nil {
		return invite, err
	}
	if len(existing) == 0 {
		return invite, ErrInviteNotFound
	}
	return existing[0], nil
}

// GetPublicInfo never errors on an invalid/expired/used token -- see
// parentinvitespg.PublicInfoResponse's doc comment.
func GetPublicInfo(ctx context.Context, token string) (res parentinvitespg.PublicInfoResponse, err error) {
	ctx = logger.Enter(ctx, "parentinvites.GetPublicInfo", token)
	defer logger.Exit(ctx, "parentinvites.GetPublicInfo", &res, &err)

	invite, err := getByToken(ctx, token)
	if err != nil {
		if err == ErrInviteNotFound {
			return parentinvitespg.PublicInfoResponse{Usable: false, Reason: "Nieprawidłowy link zaproszenia"}, nil
		}
		return res, err
	}

	if !invite.IsUsable(time.Now()) {
		reason := "To zaproszenie wygasło"
		if invite.ConsumedByAccountId != "" {
			reason = "To zaproszenie zostało już wykorzystane"
		}
		return parentinvitespg.PublicInfoResponse{Usable: false, Reason: reason}, nil
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

	return parentinvitespg.PublicInfoResponse{InviterName: inviterName, Usable: true}, nil
}
