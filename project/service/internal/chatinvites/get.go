package chatinvites

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/logger"
	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	chatinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/chatinvites"
	"github.com/Johncode248/Rodzic-Chat/project/service/internal/accounts"
)

var ErrInviteNotFound = errors.New("invite not found")

func byPinHash(pinHash string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "pin_hash", Operator: "=", Value: pinHash}}},
	}}
}

func getByPin(ctx context.Context, pin string) (invite chatinvitespg.ChatInvite, err error) {
	existing, err := gormdb.Select[chatinvitespg.ChatInvite](ctx, byPinHash(pinhash.HashPin(pin)))
	if err != nil {
		return invite, err
	}
	if len(existing) == 0 {
		return invite, ErrInviteNotFound
	}
	return existing[0], nil
}

// Lookup resolves a PIN for an already-authenticated caller, before they
// commit to Accept. Never errors on an invalid/expired/used PIN --
// Usable=false + Reason renders more easily on a client than a raw HTTP
// error. Does not consume the invite.
func Lookup(ctx context.Context, pin string) (res chatinvitespg.LookupResponse, err error) {
	ctx = logger.Enter(ctx, "chatinvites.Lookup", "<redacted>")
	defer logger.Exit(ctx, "chatinvites.Lookup", &res, &err)

	invite, err := getByPin(ctx, pin)
	if err != nil {
		if err == ErrInviteNotFound {
			return chatinvitespg.LookupResponse{Usable: false, Reason: "Nieprawidłowy kod zaproszenia"}, nil
		}
		return res, err
	}

	if !invite.IsUsable(time.Now()) {
		reason := "To zaproszenie wygasło"
		if invite.ConsumedByAccountId != "" {
			reason = "To zaproszenie zostało już wykorzystane"
		}
		return chatinvitespg.LookupResponse{Usable: false, Reason: reason}, nil
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

	return chatinvitespg.LookupResponse{InviterName: inviterName, Usable: true}, nil
}
