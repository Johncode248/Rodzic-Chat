package parentinvites

import (
	"context"
	"errors"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
)

var ErrInviteNotFound = errors.New("invite not found")
var ErrInviteNotUsable = errors.New("this invite is expired or already used")

func byPinHash(pinHash string) databasepg.Query {
	return databasepg.Query{Limit: 1, Groups: []databasepg.Group{
		{Filters: []databasepg.Filter{{Key: "pin_hash", Operator: "=", Value: pinHash}}},
	}}
}

func getByPin(ctx context.Context, pin string) (invite parentinvitespg.ParentInvite, err error) {
	existing, err := gormdb.Select[parentinvitespg.ParentInvite](ctx, byPinHash(pinhash.HashPin(pin)))
	if err != nil {
		return invite, err
	}
	if len(existing) == 0 {
		return invite, ErrInviteNotFound
	}
	return existing[0], nil
}
