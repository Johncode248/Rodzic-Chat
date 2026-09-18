package devicetokens

import (
	"context"
	"errors"
	"time"

	databasepg "github.com/Johncode248/Rodzic-Chat/database/common"
	"github.com/Johncode248/Rodzic-Chat/database/gormdb"
	devicetokenspg "github.com/Johncode248/Rodzic-Chat/project/common/devicetokens"
	"github.com/google/uuid"
)

var ErrTokenRequired = errors.New("token is required")
var ErrInvalidPlatform = errors.New("platform must be 'android' or 'ios'")

// Register upserts the caller's device token for push notifications. Keyed
// on the token itself (unique per physical install), not on account: if
// the same device previously registered under a different account (log
// out, log back in as someone else on a shared/re-used device), this
// repoints the existing row at accountId instead of erroring, so the old
// account stops getting pushes meant for whoever's using the device now.
func Register(ctx context.Context, accountId string, req devicetokenspg.RegisterRequest) (res devicetokenspg.DeviceToken, err error) {
	if req.Token == "" {
		return res, ErrTokenRequired
	}
	platform := devicetokenspg.Platform(req.Platform)
	if !platform.Valid() {
		return res, ErrInvalidPlatform
	}

	now := time.Now()

	existing, err := gormdb.Select[devicetokenspg.DeviceToken](ctx, databasepg.Query{
		Limit: 1,
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "token", Operator: "=", Value: req.Token}}},
		},
	})
	if err != nil {
		return res, err
	}

	if len(existing) > 0 {
		updated := existing[0]
		updated.AccountId = accountId
		updated.Platform = platform
		updated.UpdateTime = now

		if err = gormdb.Update(ctx, updated, databasepg.Query{
			Groups: []databasepg.Group{
				{Filters: []databasepg.Filter{{Key: "id", Operator: "=", Value: updated.Id}}},
			},
		}); err != nil {
			return res, err
		}
		return updated, nil
	}

	created := devicetokenspg.DeviceToken{
		Id:         uuid.NewString(),
		AccountId:  accountId,
		Platform:   platform,
		Token:      req.Token,
		CreateTime: now,
		UpdateTime: now,
	}
	if err = gormdb.Create(ctx, &created); err != nil {
		return res, err
	}
	return created, nil
}

// Unregister removes a device token, scoped to the caller's own account --
// called on logout so a signed-out device stops receiving pushes for an
// account it's no longer signed into.
func Unregister(ctx context.Context, accountId, token string) error {
	if token == "" {
		return ErrTokenRequired
	}
	_, err := gormdb.Delete[devicetokenspg.DeviceToken](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{
				{Key: "token", Operator: "=", Value: token},
				{Key: "account_id", Operator: "=", Value: accountId},
			}},
		},
	})
	return err
}

// ForAccounts returns every device token registered to any of the given
// accounts -- the fan-out list for a single push send.
func ForAccounts(ctx context.Context, accountIds []string) ([]devicetokenspg.DeviceToken, error) {
	if len(accountIds) == 0 {
		return nil, nil
	}
	return gormdb.Select[devicetokenspg.DeviceToken](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "account_id", Operator: "IN", Value: accountIds}}},
		},
	})
}

// DeleteTokens hard-deletes device tokens by their id -- used to prune
// tokens FCM reports as no-longer-valid (app uninstalled, token rotated
// out from under us) after a send attempt.
func DeleteTokens(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := gormdb.Delete[devicetokenspg.DeviceToken](ctx, databasepg.Query{
		Groups: []databasepg.Group{
			{Filters: []databasepg.Filter{{Key: "id", Operator: "IN", Value: ids}}},
		},
	})
	return err
}
