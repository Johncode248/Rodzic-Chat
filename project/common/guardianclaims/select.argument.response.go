package guardianclaimspg

type SelectResponse struct {
	GuardianClaims []GuardianClaim `json:"guardian_claims,omitempty"`
}
