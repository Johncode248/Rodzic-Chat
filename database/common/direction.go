package databasepg

type Direction int

const (
	DirectionUnknown Direction = iota
	DirectionAscending
	DirectionDescending
)

func (s Direction) String() string {
	switch s {
	case DirectionUnknown:
		return "Unknown"
	case DirectionAscending:
		return "ASC"
	case DirectionDescending:
		return "DESC"
	default:
		return "Unknown"
	}
}

func (s Direction) IsAscending() bool {
	return s == DirectionAscending
}
