package mutations

type Transitional[F, T any] struct {
	Foundation F
	Transient  T
}
