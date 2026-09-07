package organizationspg

// childrenLimitMonthlyPriceGrosze maps each selectable capacity tier to its
// monthly platform subscription price in grosze (1 zl = 100 grosze). This is
// what a placowka pays Rodzic Chat by bank transfer for using the platform --
// unrelated to the per-payment Stripe fee parents pay when paying czesne/
// obiady online (see project/service/internal/payments).
var childrenLimitMonthlyPriceGrosze = map[int]int64{
	20:  6900,
	30:  8900,
	40:  10900,
	50:  12900,
	80:  16900,
	100: 19900,
	120: 22900,
	200: 27900,
	300: 34900,
	400: 42900,
	500: 51900,
}

// childrenLimitOptions is the fixed, ascending list of selectable capacity
// tiers -- also the source of truth for the "oneof" validation on
// ChildrenLimit fields below. 500 is the hard ceiling: an organization
// needing more must be handled manually (custom pricing), not through
// self-service registration/settings.
var childrenLimitOptions = []int{20, 30, 40, 50, 80, 100, 120, 200, 300, 400, 500}

// ChildrenLimitOptions returns the selectable capacity tiers, ascending.
func ChildrenLimitOptions() []int {
	options := make([]int, len(childrenLimitOptions))
	copy(options, childrenLimitOptions)
	return options
}

// IsValidChildrenLimit reports whether limit is one of ChildrenLimitOptions().
func IsValidChildrenLimit(limit int) bool {
	_, ok := childrenLimitMonthlyPriceGrosze[limit]
	return ok
}

// MonthlyPriceGrosze returns the monthly subscription price for a given
// capacity tier. ok is false if limit isn't one of ChildrenLimitOptions().
func MonthlyPriceGrosze(limit int) (grosze int64, ok bool) {
	grosze, ok = childrenLimitMonthlyPriceGrosze[limit]
	return grosze, ok
}
