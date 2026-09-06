package groupspg

type CreateRequest struct {
	Name                              string `json:"name,omitempty"`
	TeacherId                         string `json:"teacher_id,omitempty"`
	AgeRange                          string `json:"age_range,omitempty"`
	RequireChildBirthDateVerification bool   `json:"require_child_birth_date_verification,omitempty"`
}
