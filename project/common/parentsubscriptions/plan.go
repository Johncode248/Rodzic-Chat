package parentsubscriptionspg

// Plan is a parent's AI subscription tier -- entirely separate from
// project/common/organizations' SubscriptionStatus (a placowka paying for
// the platform itself). Every parent account has exactly one
// ParentSubscription row, defaulting to PlanFree the first time it's read
// (see project/service/internal/parentsubscriptions.GetOrCreate).
type Plan int

const (
	PlanUnknown Plan = iota
	PlanFree
	PlanPro
	PlanMax
)

func (p Plan) String() string {
	switch p {
	case PlanFree:
		return "Free"
	case PlanPro:
		return "Pro"
	case PlanMax:
		return "Max"
	default:
		return "Unknown"
	}
}

func (p Plan) Ordinal() int {
	return int(p)
}

func MapPlan(value string) Plan {
	switch value {
	case "free":
		return PlanFree
	case "pro":
		return PlanPro
	case "max":
		return PlanMax
	default:
		return PlanUnknown
	}
}

// monthlyAiRequestLimit is a PLACEHOLDER -- the product owner hasn't
// decided real numbers yet (as of this writing). -1 means unlimited.
// Change freely; nothing else needs updating in lockstep.
var monthlyAiRequestLimit = map[Plan]int64{
	PlanFree: 5,
	PlanPro:  200,
	PlanMax:  -1,
}

// MonthlyAiRequestLimit returns the plan's monthly AI request allowance,
// or -1 for unlimited (PlanMax). Unknown plans get PlanFree's limit.
func MonthlyAiRequestLimit(plan Plan) int64 {
	if limit, ok := monthlyAiRequestLimit[plan]; ok {
		return limit
	}
	return monthlyAiRequestLimit[PlanFree]
}

// monthlyAiTokenLimit is the REAL enforcement cap on the AI assistant (see
// ParentSubscription.AiTokensUsed's doc comment for why tokens, not requests). Deliberately kept
// small for PlanFree -- MVP ships with only PlanFree actually selectable (see PlanPro/PlanMax
// being disabled in the client's Plan screen), and every Claude token has a real, non-trivial
// cost with no revenue behind it yet. 20,000 tokens/month is roughly 6-10 short tool-use
// conversations; tune freely, nothing else needs updating in lockstep.
var monthlyAiTokenLimit = map[Plan]int64{
	PlanFree: 20_000,
	PlanPro:  400_000,
	PlanMax:  -1,
}

// MonthlyAiTokenLimit returns the plan's monthly AI token allowance, or -1 for unlimited
// (PlanMax). Unknown plans get PlanFree's limit.
func MonthlyAiTokenLimit(plan Plan) int64 {
	if limit, ok := monthlyAiTokenLimit[plan]; ok {
		return limit
	}
	return monthlyAiTokenLimit[PlanFree]
}

// DailyGooglePlacesSearchLimit caps how many Google Places searches ONE parent can trigger per
// day, regardless of plan (MVP has only PlanFree active) -- a per-parent throttle sitting
// underneath the global MonthlyGooglePlacesRequestBudget (see project/common/places), so one
// very active parent can't burn through the whole platform's shared monthly budget alone.
const DailyGooglePlacesSearchLimit = 15

// MaxRecommendationsPerParent caps how many distinct places one parent account can recommend in
// total (not per period -- lifetime) -- see project/service/internal/places.Recommend. Keeps the
// "poleć miejsce" feature from being spammable while still comfortably covering a genuine
// parent's real recommendations.
const MaxRecommendationsPerParent = 50
