package payments

import platformsettingspg "github.com/Johncode248/Rodzic-Chat/project/common/platformsettings"

// CheckoutAmounts is the result of ComputeCheckoutAmounts: what a parent
// must pay online so the school still receives exactly baseAmountGrosze
// after Stripe's own cut and the platform's fee.
type CheckoutAmounts struct {
	PlatformFeeGrosze       int64
	StripeFeeEstimateGrosze int64
	TotalAmountGrosze       int64
}

// ComputeCheckoutAmounts solves for the total a parent must be charged so
// that, after Stripe's percentage+fixed cut on the *total*, what's left
// still covers baseAmountGrosze (what the school must receive) plus the
// platform's own flat fee:
//
//	total - stripeFixed - total*stripePercent = base + platformFee
//	total * (1 - stripePercent) = base + platformFee + stripeFixed
//	total = (base + platformFee + stripeFixed) / (1 - stripePercent)
//
// This is only ever an *estimate* of Stripe's real fee for whatever card/
// method the parent actually uses -- the connected account's transfer is
// separately pinned to exactly baseAmountGrosze (see
// project/service/internal/payments/checkout.go), so any gap between this
// estimate and Stripe's real fee only ever changes the platform's own
// margin on that payment, never what the school receives or what the
// parent was quoted before paying.
func ComputeCheckoutAmounts(baseAmountGrosze int64, settings platformsettingspg.PlatformSettings) CheckoutAmounts {
	platformFee := settings.PlatformFeeGrosze
	stripeFixed := settings.StripeAssumedFixedGrosze
	stripePercentBps := settings.StripeAssumedPercentBps // e.g. 150 = 1.5%

	numerator := baseAmountGrosze + platformFee + stripeFixed

	// total = numerator / (1 - pct). Working in integer basis points
	// (10000 = 100%) and rounding the division up so the school is never
	// short by a grosz: total = ceil(numerator * 10000 / (10000 - bps)).
	denominatorBps := int64(10000) - stripePercentBps
	totalAmount := ceilDiv(numerator*10000, denominatorBps)

	stripeFeeEstimate := totalAmount - baseAmountGrosze - platformFee

	return CheckoutAmounts{
		PlatformFeeGrosze:       platformFee,
		StripeFeeEstimateGrosze: stripeFeeEstimate,
		TotalAmountGrosze:       totalAmount,
	}
}

func ceilDiv(a, b int64) int64 {
	if b == 0 {
		return a
	}
	if a%b == 0 {
		return a / b
	}
	return a/b + 1
}
