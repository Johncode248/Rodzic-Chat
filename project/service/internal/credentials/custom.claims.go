package credentials

import "github.com/golang-jwt/jwt/v4"

type MyCustomClaims struct {
	UserId    string `json:"userId"`
	SessionId string `json:"sessionId"`
	Email     string `json:"email"`
	// OrganizationId is the *active* organization for this token -- every
	// existing consumer (AuthMiddleware -> context -> AddOrganizationAuth)
	// keeps its exact current meaning and needs zero changes.
	OrganizationId string `json:"organizationId"`
	// OrganizationIds is every organization the account currently belongs
	// to (via accountorganizationspg.AccountOrganization). Purely additive:
	// only the "list my organizations" / "switch active organization"
	// endpoints read it. A single-org account (director, teacher, or a
	// parent who never joined a second facility) always has exactly one
	// entry here, equal to OrganizationId.
	OrganizationIds []string `json:"organizationIds,omitempty"`
	AccountType     int      `json:"accountType"`
	TokenKind       string   `json:"tokenKind"`
	jwt.RegisteredClaims
}
