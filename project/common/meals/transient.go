package mealspg

type Transient struct {
	Previous MealCancellation
	Current  MealCancellation
	Future   MealCancellation
}

func (t Transient) Update() Transient {
	return Transient{
		Previous: t.Current,
		Current:  t.Future,
		Future:   MealCancellation{},
	}
}
