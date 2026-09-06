package authorizationpg

import (
	"errors"

	accountspg "github.com/Johncode248/Rodzic-Chat/project/common/accounts"
	organizationspg "github.com/Johncode248/Rodzic-Chat/project/common/organizations"
)

func MapAccountType(value string) (accountspg.AccountType, error) {
	switch value {
	case "director":
		return accountspg.AccountTypeDirector, nil
	case "co_director":
		return accountspg.AccountTypeCoDirector, nil
	case "parent":
		return accountspg.AccountTypeParent, nil
	case "teacher":
		return accountspg.AccountTypeTeacher, nil
	default:
		return accountspg.AccountTypeUnknown, errors.New("unrecognized account_type")
	}
}

func MapOrganizationType(value string) (organizationspg.OrganizationType, error) {
	switch value {
	case "kindergarten":
		return organizationspg.OrganizationTypeKindergarten, nil
	case "nursery":
		return organizationspg.OrganizationTypeNursery, nil
	case "school":
		return organizationspg.OrganizationTypeSchool, nil
	case "extracurricular":
		return organizationspg.OrganizationTypeExtracurricular, nil
	case "independent":
		return organizationspg.OrganizationTypeIndependent, nil
	default:
		return organizationspg.OrganizationTypeUnknown, errors.New("unrecognized organization_type")
	}
}
