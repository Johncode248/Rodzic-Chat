package groupspg

type CreateRequest struct {
	Name                              string `json:"name,omitempty"`
	TeacherId                         string `json:"teacher_id,omitempty"`
	RequireChildBirthDateVerification bool   `json:"require_child_birth_date_verification,omitempty"`
}
