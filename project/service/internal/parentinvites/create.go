package parentinvites

import (
	"context"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/library/pinhash"
	"github.com/Johncode248/Rodzic-Chat/library/random"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentinvitespg "github.com/Johncode248/Rodzic-Chat/project/common/parentinvites"
	"github.com/google/uuid"
)

const expiryDuration = 72 * time.Hour

// maxPinGenerationAttempts bounds retrying GeneratePin on the rare
// uniqueIndex collision on pin_hash.
const maxPinGenerationAttempts = 5

// Create issues a fresh invite for accountId ("rodzic nr 1") to share with
// a co-parent. No limit on how many outstanding invites one account can
// have -- creating a new one doesn't invalidate older ones. Returns the
// plaintext PIN alongside the stored row -- it's never retrievable again.
func Create(ctx context.Context, accountId string) (invite parentinvitespg.ParentInvite, plaintextPin string, err error) {
	ctx = logger.Enter(ctx, "parentinvites.Create", accountId)
	defer logger.Exit(ctx, "parentinvites.Create", &invite, &err)

	now := time.Now()

	for attempt := 0; ; attempt++ {
		plaintextPin = random.GeneratePin()
		invite = parentinvitespg.ParentInvite{
			Id:               uuid.NewString(),
			PinHash:          pinhash.HashPin(plaintextPin),
			InviterAccountId: accountId,
			ExpiresAt:        now.Add(expiryDuration),
			CreateTime:       now,
		}

		err = gormdb.Create(ctx, &invite)
		if err == nil {
			break
		}
		if !pinhash.IsUniqueViolation(err) || attempt >= maxPinGenerationAttempts-1 {
			return parentinvitespg.ParentInvite{}, "", err
		}
	}

	return invite, plaintextPin, nil
}
