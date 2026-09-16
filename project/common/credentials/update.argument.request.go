package credentialspg

type UpdateRequest struct {
	//Query    databasepg.Query `json:"query,omitempty"`
	Password *string `json:"password,omitempty"`
	// CurrentPassword must match the account's existing password before a
	// change is accepted -- see credentialsgt.Update. Not required when the
	// account has no Credential row yet (OAuth-only account setting a
	// password for the first time), since there is nothing to verify
	// against.
	CurrentPassword *string `json:"current_password,omitempty"`
}
