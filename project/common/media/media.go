package mediapg

import "time"

type Media struct {
	Id             string `json:"id,omitempty" gorm:"primaryKey"`
	OrganizationId string `json:"organization_id,omitempty" gorm:"not null;index"`
	EventId        string `json:"event_id,omitempty" gorm:"index"`
	GroupId        string `json:"group_id,omitempty" gorm:"index"`
	UploaderId     string `json:"uploader_id,omitempty" gorm:"not null"`
	ObjectKey      string `json:"object_key,omitempty" gorm:"not null"`
	// OriginalFilename/ContentType are what the uploader's browser sent
	// (e.g. "zgoda-wycieczka-zoo.pdf", "application/pdf") -- ObjectKey
	// itself is just a UUID path, not something fit to show a downloader.
	OriginalFilename string `json:"original_filename,omitempty"`
	ContentType      string `json:"content_type,omitempty"`
	Caption          string `json:"caption,omitempty"`
	// Category free-text tags what kind of document this is for the
	// mobile Gallery tab to filter/label by, e.g. "schedule" (plan zajęć),
	// "meal_plan" (plan obiadowy), or "" for a plain gallery photo. Purely
	// a display hint -- the backend doesn't validate or branch on it.
	Category   string    `json:"category,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
}

// MediaWithUrl is Media plus a resolved, time-limited download URL -- the
// shape actually returned from the gallery select endpoint instead of the
// raw row (the client can't do anything useful with a bare ObjectKey).
type MediaWithUrl struct {
	Media Media  `json:"media,omitempty"`
	Url   string `json:"url,omitempty"`
}
