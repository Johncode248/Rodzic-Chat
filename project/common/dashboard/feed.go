package dashboardpg

import "time"

// FeedItem is one entry in the director dashboard's activity feed --
// deliberately a flat, tagged-union shape (no dedicated DB table) so the
// feed can be assembled at read time from already-existing sources
// (new parent registrations, absence reports, chat announcements) instead
// of requiring a new audit-log write path everywhere something happens.
type FeedItem struct {
	Type       string    `json:"type,omitempty"` // "new_parent" | "absence" | "announcement"
	Id         string    `json:"id,omitempty"`
	Title      string    `json:"title,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

type FeedResponse struct {
	Items []FeedItem `json:"items,omitempty"`
}
