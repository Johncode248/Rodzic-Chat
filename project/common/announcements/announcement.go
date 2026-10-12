package announcementspg

import "time"

// Announcement is a short developer-authored note shown in the parent-facing "Nowości" tab --
// v1 ships with no placówka marketplace, so this is the one channel for the developer to tell
// parents about other things (a sibling app, a seasonal partner discount, ...). Rows are written
// directly in the database by the developer, there is no in-app authoring UI. Every announcement
// opens its own in-app detail screen when tapped; [LinkUrl] (optional, "" = none) adds a "Zobacz
// ofertę" button there that opens that URL externally, and [ImageUrl] (optional) adds a hero
// photo to both the list card and the detail screen.
type Announcement struct {
	Id          string `json:"id,omitempty" gorm:"primaryKey"`
	Title       string `json:"title,omitempty" gorm:"not null"`
	Description string `json:"description,omitempty" gorm:"not null"`
	LinkUrl     string `json:"link_url,omitempty"`
	// ImageUrl is optional -- a photo for the detail screen's hero banner (e.g. actual Christmas
	// trees for a tree discount, or the sibling app's icon for an app announcement). "" shows no
	// image, just the icon badge the list/detail already fall back to.
	ImageUrl string `json:"image_url,omitempty"`
	// Category groups announcements for the filter row on the "Info" tab: "info" (general
	// information), "sale" (sales / discounts) or "event" (events). Anything else is shown as
	// "other" by the clients. New rows default to "info".
	Category   string    `json:"category,omitempty" gorm:"not null;default:info"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null;index"`
}

type ListResponse struct {
	Announcements []Announcement `json:"announcements"`
}
