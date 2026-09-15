package parentsubscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
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

	current, err := GetOrCreate(ctx, accountId)
	if err != nil {
		return res, err
	}

	current.AiRequestsUsed++
	current.AiTokensUsed += tokens
	current.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &current, byAccountId(accountId)); err != nil {
		return res, err
	}

	return current, nil
}
