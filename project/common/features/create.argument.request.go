package featurespg

type CreateRequest struct {
	Feature Feature `json:"feature,omitempty"`
	Enabled bool    `json:"enabled"`
}
