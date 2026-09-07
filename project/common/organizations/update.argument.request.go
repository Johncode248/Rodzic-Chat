package organizationspg

import databasepg "github.com/Johncode248/Rodzic-Chat/database/common"

type UpdateRequest struct {
	Query         databasepg.Query `json:"query,omitempty"`
	Name          *string          `json:"name,omitempty"`
	Address       *string          `json:"address,omitempty"`
	Nip           *string          `json:"nip,omitempty"`
	LogoUrl       *string          `json:"logo_url,omitempty"`
	ChildrenLimit *int             `json:"children_limit,omitempty"`
	PrimaryColor  *string          `json:"primary_color,omitempty"`
}
