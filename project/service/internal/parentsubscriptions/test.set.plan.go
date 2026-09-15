package parentsubscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	"github.com/Johncode248/Rodzic-Chat/logger"
	parentsubscriptionspg "github.com/Johncode248/Rodzic-Chat/project/common/parentsubscriptions"
)

var ErrUnknownPlan = errors.New("unknown plan -- expected free, pro, or max")

// TestSetPlan changes accountId's plan directly, with no payment involved
// at all. This exists ONLY so the app's plan/limit UI can be built and
// verified before real billing (StoreKit/Play Billing) is wired up --
// Apple/Google require in-app purchase for exactly this kind of digital
// entitlement, so this must never be reachable from a shipped production
// build. See transport/parentsubscriptions/test.set.plan.go for the gate.
func TestSetPlan(ctx context.Context, accountId string, plan parentsubscriptionspg.Plan) (res parentsubscriptionspg.ParentSubscription, err error) {
	ctx = logger.Enter(ctx, "parentsubscriptions.TestSetPlan", accountId, plan)
	defer logger.Exit(ctx, "parentsubscriptions.TestSetPlan", &res, &err)

	if plan == parentsubscriptionspg.PlanUnknown {
		return res, ErrUnknownPlan
	}

	current, err := GetOrCreate(ctx, accountId)
	if err != nil {
		return res, err
	}

	current.Plan = plan
	current.UpdateTime = time.Now()

	if err = gormdb.Update(ctx, &current, byAccountId(accountId)); err != nil {
		return res, err
	}

	return current, nil
}
