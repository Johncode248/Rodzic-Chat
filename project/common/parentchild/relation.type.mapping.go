package parentchildpg

func MapRelationType(value string) RelationType {
	switch value {
	case "mother":
		return RelationTypeMother
	case "father":
		return RelationTypeFather
	case "guardian":
		return RelationTypeGuardian
	default:
		return RelationTypeUnknown
	}
}
