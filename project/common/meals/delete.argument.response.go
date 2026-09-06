package mealspg

type DeleteResponse struct {
	MealCancellations []MealCancellation `json:"meal_cancellations,omitempty"`
}
