package accountspg

type SelectResponse struct {
	Accounts []Account `json:"accounts,omitempty"`
}
