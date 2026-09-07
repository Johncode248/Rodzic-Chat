package organizationspg

// SubscriptionOverrideRequest is the platform-operator-only action of
// exempting an organization from the subscription fee entirely (or
// revoking that exemption) -- see
// project/service/transport/organizations/subscription.override.go for the
// shared-secret gate, and SubscriptionFreeOverride on Organization.
type SubscriptionOverrideRequest struct {
	OrganizationId string `json:"organization_id,omitempty"`
	FreeOverride   bool   `json:"free_override"`
}

type SubscriptionOverrideResponse struct {
	Organization Organization `json:"organization,omitempty"`
}
