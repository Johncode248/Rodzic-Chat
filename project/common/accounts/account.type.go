package accountspg

type AccountType int

// AccountTypeCoDirector must always be appended after the last existing
// value, never inserted -- this is stored as a plain int ordinal in the DB,
// so inserting a new value in the middle would silently reinterpret every
// existing row's account type.
const (
	AccountTypeUnknown AccountType = iota
	AccountTypeParent
	AccountTypeTeacher
	AccountTypeDirector
	AccountTypeCoDirector
)

func (a AccountType) String() string {
	switch a {
	case AccountTypeParent:
		return "Parent"
	case AccountTypeTeacher:
		return "Teacher"
	case AccountTypeDirector:
		return "Director"
	case AccountTypeCoDirector:
		return "CoDirector"
	default:
		return "Unknown"
	}
}

func (a AccountType) Ordinal() int {
	return int(a)
}
