package organizationspg

type OrganizationType int

const (
	OrganizationTypeUnknown OrganizationType = iota
	OrganizationTypeKindergarten
	OrganizationTypeNursery
	OrganizationTypeSchool
	OrganizationTypeExtracurricular
	OrganizationTypeIndependent
)

func (o OrganizationType) String() string {
	switch o {
	case OrganizationTypeKindergarten:
		return "Kindergarten"
	case OrganizationTypeNursery:
		return "Nursery"
	case OrganizationTypeSchool:
		return "School"
	case OrganizationTypeExtracurricular:
		return "Extracurricular"
	case OrganizationTypeIndependent:
		return "Independent"
	default:
		return "Unknown"
	}
}

func (o OrganizationType) Ordinal() int {
	return int(o)
}
