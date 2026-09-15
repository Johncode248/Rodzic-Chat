package parentinvites

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
)

var ErrCannotAcceptOwnInvite = errors.New("you cannot accept your own invite")

// Accept links an already-authenticated account to every child the
// inviter parents -- the path for a co-parent who already has a Rodzic
// Chat account (see Register for the no-account-yet path).
func Accept(ctx context.Context, token string, accepterAccountId string) (res parentinvitespg.AcceptResponse, err error) {
	ctx = logger.Enter(ctx, "parentinvites.Accept", token, accepterAccountId)
	defer logger.Exit(ctx, "parentinvites.Accept", &res, &err)

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

	return linkToInviterChildren(ctx, invite, accepterAccountId)
}
