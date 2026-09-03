package credentialspg

type UpdateRequest struct {
	//Query    databasepg.Query `json:"query,omitempty"`
	Password *string `json:"password,omitempty"`
}
