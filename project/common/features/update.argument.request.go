package featurespg

type UpdateRequest struct {
	Feature Feature `json:"feature,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}
