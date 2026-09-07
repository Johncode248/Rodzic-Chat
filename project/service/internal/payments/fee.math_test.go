package payments

import (
	"testing"

	platformsettingspg "github.com/Johncode248/Rodzic-Chat/project/common/platformsettings"
)

func TestComputeCheckoutAmounts_SchoolAlwaysReceivesExactlyBase(t *testing.T) {
	settings := platformsettingspg.PlatformSettings{
		PlatformFeeGrosze:        100, // 1 zł
		StripeAssumedPercentBps:  150, // 1.5%
		StripeAssumedFixedGrosze: 100, // 1 zł
	}

	cases := []int64{100, 999, 5000, 50000, 123456, 1}

	for _, base := range cases {
		amounts := ComputeCheckoutAmounts(base, settings)

		if amounts.TotalAmountGrosze <= base {
			t.Fatalf("base=%d: total (%d) must exceed base", base, amounts.TotalAmountGrosze)
		}

		// The whole point of this formula: after subtracting the assumed
		// Stripe cut from the total, what's left must cover at least
		// base+platformFee (never less -- a shortfall would mean the
		// transfer_data.amount=base pin in checkout.go silently eats into
		// the platform's fee or worse).
		assumedStripeFee := amounts.TotalAmountGrosze*settings.StripeAssumedPercentBps/10000 + settings.StripeAssumedFixedGrosze
		remainder := amounts.TotalAmountGrosze - assumedStripeFee
		want := base + settings.PlatformFeeGrosze
		if remainder < want {
			t.Fatalf("base=%d: remainder after Stripe's cut (%d) is less than base+platformFee (%d)", base, remainder, want)
		}
		// And it shouldn't overshoot by more than a grosz or two of
		// rounding slack either -- a parent should never be visibly
		// overcharged.
		if remainder-want > 2 {
			t.Fatalf("base=%d: remainder (%d) overshoots base+platformFee (%d) by more than rounding slack", base, remainder, want)
		}
	}
}

func TestComputeCheckoutAmounts_ZeroFees(t *testing.T) {
	settings := platformsettingspg.PlatformSettings{
		PlatformFeeGrosze:        0,
		StripeAssumedPercentBps:  0,
		StripeAssumedFixedGrosze: 0,
	}

	amounts := ComputeCheckoutAmounts(50000, settings)
	if amounts.TotalAmountGrosze != 50000 {
		t.Fatalf("with zero fees, total should equal base exactly; got %d", amounts.TotalAmountGrosze)
	}
}
