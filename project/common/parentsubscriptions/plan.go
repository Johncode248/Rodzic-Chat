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

// monthlyAiRequestLimit/monthlyAiTokenLimit are DISPLAY-ONLY for MVP -- the AI assistant endpoint
// itself is hard-disabled server-side (see project/service/transport/aichat.Send) and
// ParentSubscription no longer tracks usage against these at all (see its doc comment), so
// nothing enforces them any more. Kept only so the "Limity" screen has a number to show; change
// freely, nothing else needs updating in lockstep.
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

// DailyGooglePlacesSearchLimit is also DISPLAY-ONLY for MVP now -- the per-parent daily counter
// it used to gate (ParentSubscription.GooglePlacesUsedToday) was removed; only the shared global
// monthly budget (placespg.MonthlyGooglePlacesRequestBudget) still actually gates Google Places
// calls (see project/service/internal/places.SearchWithFallback). One very active parent could
// now exhaust that shared budget alone -- re-add a per-parent counter if that turns out to matter.
const DailyGooglePlacesSearchLimit = 15

// MaxRecommendationsPerParent caps how many distinct places one parent account can recommend in
// total (not per period -- lifetime) -- see project/service/internal/places.Recommend. Keeps the
// "poleć miejsce" feature from being spammable while still comfortably covering a genuine
// parent's real recommendations.
const MaxRecommendationsPerParent = 50
