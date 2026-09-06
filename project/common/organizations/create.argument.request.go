package organizationspg

type CreateRequest struct {
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	Subdomain    string `json:"subdomain,omitempty"`
	Address      string `json:"address,omitempty"`
	Nip          string `json:"nip,omitempty"`
	LogoUrl      string `json:"logo_url,omitempty"`
	PrimaryColor string `json:"primary_color,omitempty"`
}
