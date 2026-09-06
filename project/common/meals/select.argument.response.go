package mealspg

type SelectResponse struct {
	MealCancellations []MealCancellation `json:"meal_cancellations,omitempty"`
}
