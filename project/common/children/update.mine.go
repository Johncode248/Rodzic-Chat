package childrenpg

// UpdateMineRequest is a parent editing their own child's details --
// first/last name, birth date, notes, and PESEL are all optional partial
// updates (nil = leave unchanged). Pesel, if given, is the PLAINTEXT
// value; project/service/internal/children.UpdateMine encrypts it before
// storage (see library/crypto) -- it is never accepted or returned as
// ciphertext over the wire.
type UpdateMineRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	// BirthDate as "2006-01-02".
	BirthDate *string `json:"birth_date,omitempty"`
	Notes     *string `json:"notes,omitempty"`
	Pesel     *string `json:"pesel,omitempty"`
}

type RevealPeselResponse struct {
	Pesel string `json:"pesel,omitempty"`
}
