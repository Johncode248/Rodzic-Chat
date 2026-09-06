package parentchildpg

type RelationType int

const (
	RelationTypeUnknown RelationType = iota
	RelationTypeMother
	RelationTypeFather
	RelationTypeGuardian
)

func (r RelationType) String() string {
	switch r {
	case RelationTypeMother:
		return "Mother"
	case RelationTypeFather:
		return "Father"
	case RelationTypeGuardian:
		return "Guardian"
	default:
		return "Unknown"
	}
}

func (r RelationType) Ordinal() int {
	return int(r)
}
