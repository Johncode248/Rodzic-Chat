package featurespg

type SelectResponse struct {
	OrganizationFeatures []OrganizationFeature `json:"organization_features,omitempty"`
}
