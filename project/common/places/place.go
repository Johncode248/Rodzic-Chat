package placespg

import "time"

// Place is a real-world location or service a parent can recommend
// (project/service/internal/places). It is never scraped or fetched in
// bulk from OpenStreetMap -- a row is created lazily, the first time some
// parent recommends that place (either picked from an OpenStreetMap
// Nominatim search result, or typed in by hand when OSM doesn't have it).
// From then on this row -- not OSM -- is the source of truth for display;
// Latitude/Longitude/Address are a point-in-time copy taken at creation.
//
// Source records where the initial data came from ("osm" or "manual"),
// and ExternalRef holds OpenStreetMap's own stable id (e.g. "node/123456")
// when Source is "osm", purely for de-duplication (see
// project/service/internal/places/recommend.go) -- it is never used to
// re-fetch anything from OSM at read time.
type Place struct {
	Id               string   `json:"id,omitempty" gorm:"primaryKey"`
	Source           string   `json:"source,omitempty" gorm:"not null"`
	ExternalRef      string   `json:"external_ref,omitempty" gorm:"index"`
	Name             string   `json:"name,omitempty" gorm:"not null"`
	Category         string   `json:"category,omitempty" gorm:"index;not null"`
	Address          string   `json:"address,omitempty"`
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
	Phone            string   `json:"phone,omitempty"`
	Website          string   `json:"website,omitempty"`
	OpeningHours     string   `json:"opening_hours,omitempty"`
	AddedByAccountId string   `json:"added_by_account_id,omitempty" gorm:"index"`

	CreateTime time.Time `json:"create_time,omitempty" gorm:"not null"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"not null"`
}

const (
	SourceOSM    = "osm"
	SourceGoogle = "google"
	SourceManual = "manual"
)

// HasCoordinates reports whether a "Pokaż na mapie" affordance makes sense
// for this place.
func (p Place) HasCoordinates() bool {
	return p.Latitude != nil && p.Longitude != nil
}
