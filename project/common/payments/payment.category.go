package paymentspg

// PaymentCategory is a display/filter hint, not something the backend
// branches on -- it just labels what a Payment is for (product direction:
// "czesne, opłata za obiady, zajęcia pozalekcyjne, opłata wycieczkowa").
type PaymentCategory int

const (
	PaymentCategoryUnknown PaymentCategory = iota
	PaymentCategoryTuition
	PaymentCategoryMeals
	PaymentCategoryExtracurricular
	PaymentCategoryTrip
	PaymentCategoryOther
)

func (c PaymentCategory) String() string {
	switch c {
	case PaymentCategoryTuition:
		return "Tuition"
	case PaymentCategoryMeals:
		return "Meals"
	case PaymentCategoryExtracurricular:
		return "Extracurricular"
	case PaymentCategoryTrip:
		return "Trip"
	case PaymentCategoryOther:
		return "Other"
	default:
		return "Unknown"
	}
}

func (c PaymentCategory) Ordinal() int {
	return int(c)
}

func MapPaymentCategory(value string) PaymentCategory {
	switch value {
	case "tuition":
		return PaymentCategoryTuition
	case "meals":
		return PaymentCategoryMeals
	case "extracurricular":
		return PaymentCategoryExtracurricular
	case "trip":
		return PaymentCategoryTrip
	case "other":
		return PaymentCategoryOther
	default:
		return PaymentCategoryUnknown
	}
}
