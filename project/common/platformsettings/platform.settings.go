package platformsettingspg

import "time"

// SingletonId is the fixed primary key of the one PlatformSettings row this
// table is ever meant to hold -- there is no per-organization override in
// v1 (see the "global default, changed by hand" product decision), so
// reads/writes always target this id rather than searching for "the" row.
const SingletonId = "platform-settings"

// PlatformSettings holds the platform-wide numbers that go into a Payment's
// online-checkout fee math (see project/common/payments). All amounts are
// in grosze (1 zł = 100 grosze) to avoid float rounding on money.
//
// PlatformFeeGrosze is Rodzic Chat's own cut per online payment (product
// direction: "np. 1 zł, ale ta kwota może być zmienna") -- changed by
// editing this row (see transport/platformsettings/update.go, guarded by a
// shared secret rather than a real superadmin role, since no such role
// exists yet).
//
// StripeAssumedPercentBps / StripeAssumedFixedGrosze are *estimates* of
// Stripe's own cut, used only to compute how much extra a parent must pay
// upfront so the school still receives exactly BaseAmount (see
// payments.ComputeCheckoutAmounts) -- the connected account's transfer is
// always pinned to the exact base amount regardless of whether these
// estimates match Stripe's real fee for the card/method actually used, so
// a mismatch here only ever affects the platform's own margin, never the
// school's payout or the parent's expectation. Update these if Stripe's
// published Polish pricing changes.
type PlatformSettings struct {
	Id                       string    `json:"id,omitempty" gorm:"primaryKey"`
	PlatformFeeGrosze        int64     `json:"platform_fee_grosze,omitempty" gorm:"not null"`
	StripeAssumedPercentBps  int64     `json:"stripe_assumed_percent_bps,omitempty" gorm:"not null"` // basis points, e.g. 150 = 1.5%
	StripeAssumedFixedGrosze int64     `json:"stripe_assumed_fixed_grosze,omitempty" gorm:"not null"`
	CreateTime               time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime               time.Time `json:"update_time,omitempty" gorm:"not null"`
}

// Defaults seed the singleton row the first time it's read and none exists
// yet -- 1 zł platform fee, 1.5% + 1 zł assumed Stripe cut (a reasonable
// starting point for EU cards; BLIK/other methods will need this tuned
// once real transactions are flowing).
func Defaults() PlatformSettings {
	now := time.Now()
	return PlatformSettings{
		Id:                       SingletonId,
		PlatformFeeGrosze:        100,
		StripeAssumedPercentBps:  150,
		StripeAssumedFixedGrosze: 100,
		CreateTime:               now,
		UpdateTime:               now,
	}
}

type UpdateRequest struct {
	PlatformFeeGrosze        *int64 `json:"platform_fee_grosze,omitempty"`
	StripeAssumedPercentBps  *int64 `json:"stripe_assumed_percent_bps,omitempty"`
	StripeAssumedFixedGrosze *int64 `json:"stripe_assumed_fixed_grosze,omitempty"`
}

type SelectResponse struct {
	PlatformSettings PlatformSettings `json:"platform_settings,omitempty"`
}

type UpdateResponse struct {
	PlatformSettings PlatformSettings `json:"platform_settings,omitempty"`
}
