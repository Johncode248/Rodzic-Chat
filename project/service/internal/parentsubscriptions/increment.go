package parentsubscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
	"gorm.io/gorm"
)

var ErrAiLimitReached = errors.New("monthly AI token limit reached for this plan")

// IncrementUsageWithTokens rolls the period forward if needed (via GetOrCreate), then bumps both
// AiRequestsUsed (by one) and AiTokensUsed (by [tokens]) in a single write, returning the row
// *after* incrementing. Called only after a successful Claude reply -- a failed call must never
// count against either counter. The actual budget check (AiTokensUsed vs.
// parentsubscriptionspg.MonthlyAiTokenLimit) happens in aichat.Send BEFORE calling Claude at all,
// not here -- tokens for a given reply aren't known until after the call completes.
func IncrementUsageWithTokens(ctx context.Context, accountId string, tokens int64) (res parentsubscriptionspg.ParentSubscription, err error) {
	ctx = logger.Enter(ctx, "parentsubscriptions.IncrementUsageWithTokens", accountId, tokens)
	defer logger.Exit(ctx, "parentsubscriptions.IncrementUsageWithTokens", &res, &err)

	// The read-modify-write sequence below (GetOrCreate, bump both
	// counters in memory, write back) has to be serialized per-account,
	// otherwise two concurrent AI replies for the same account can both
	// read the same starting counters and both write back count+1,
	// losing one increment -- under-counting usage and letting a parent
	// exceed MonthlyAiTokenLimit. pg_advisory_xact_lock keyed on
	// accountId makes every concurrent call for this account queue up one
	// at a time, mirroring the pattern already used for the per-account
	// channel limit in chatinvites.Create.
	err = gormdb.GormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", accountId).Error; err != nil {
			return err
		}

		current, err := GetOrCreate(ctx, accountId)
		if err != nil {
			return err
		}

		current.AiRequestsUsed++
		current.AiTokensUsed += tokens
		current.UpdateTime = time.Now()

		if err = gormdb.Update(ctx, &current, byAccountId(accountId)); err != nil {
			return err
		}

		res = current
		return nil
	})
	if err != nil {
		return res, err
	}

	return res, nil
}
