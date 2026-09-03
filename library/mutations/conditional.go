package mutations

type Conditional[F, C, T any] struct {
	Foundation F
	Condition  C
	Transient  T
}
