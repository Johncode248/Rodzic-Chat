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
// inviter parents -- both inviter and accepter must already have an
// account (there is no more no-account-yet path; see the removed
// register.go).
func Accept(ctx context.Context, pin string, accepterAccountId string) (res parentinvitespg.AcceptResponse, err error) {
	ctx = logger.Enter(ctx, "parentinvites.Accept", "<redacted>", accepterAccountId)
	defer logger.Exit(ctx, "parentinvites.Accept", &res, &err)

	invite, err := getByPin(ctx, pin)
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
