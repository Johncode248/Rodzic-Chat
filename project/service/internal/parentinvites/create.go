package parentinvites

import (
	"context"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/google/uuid"
)

const expiryDays = 7

// Create issues a fresh invite for accountId ("rodzic nr 1") to share with
// a co-parent. No limit on how many outstanding invites one account can
// have -- creating a new one doesn't invalidate older ones.
func Create(ctx context.Context, accountId string) (res parentinvitespg.ParentInvite, err error) {
	ctx = logger.Enter(ctx, "parentinvites.Create", accountId)
	defer logger.Exit(ctx, "parentinvites.Create", &res, &err)

	now := time.Now()
	invite := parentinvitespg.ParentInvite{
		Id:               uuid.NewString(),
		Token:            random.GenerateToken(),
		InviterAccountId: accountId,
		ExpiresAt:        now.AddDate(0, 0, expiryDays),
		CreateTime:       now,
	}
	if err = gormdb.Create(ctx, &invite); err != nil {
		return res, err
	}

	return invite, nil
}
