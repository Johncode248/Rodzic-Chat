// Package nfzpg holds the wire types for the public NFZ "API Terminy Leczenia" (Estimated
// Treatment Time API), v1.4 -- https://apinfz.nfz.gov.pl/app-itl-api-pcus/index.html. That API
// publishes NFZ's own monthly-updated wait-time forecasts (PCUS = "Prognozowany Czas Udzielenia
// Świadczenia") and provider directory for specialist consultations, hospital wards, and
// diagnostic labs contracted with the National Health Fund. It is NOT a booking system -- there
// are no live appointment slots, only a forecast in days/weeks plus the provider's contact
// details, and it does NOT cover routine primary-care ("POZ") visits like an ordinary
// pediatrician check-up, since those aren't queued/reported under this system at all. It only
// has entries for named NFZ "benefits" (specialist clinics, procedures, wards) -- see
// FindBenefit below for resolving a parent's free-text ask into one of those exact names.
package nfzpg

// Province is one of Poland's 16 "województwo" administrative regions, identified by NFZ's own
// two-digit code (as used by every province query parameter across this API).
type Province struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

// Case mirrors the NFZ "case" (przypadek) query parameter: 1 = stable (routine), 2 = urgent.
type Case int

const (
	CaseStable Case = 1
	CaseUrgent Case = 2
)

// FindBenefitRequest resolves free text (e.g. a parent's "alergolog dla dziecka") into exact
// NFZ benefit names -- the /queues endpoint requires an exact match, and NFZ's own naming is
// formal/all-caps ("PORADNIA ALERGOLOGICZNA DLA DZIECI"), so this dictionary lookup is always
// the first step before searching queues.
type FindBenefitRequest struct {
	Query string `json:"query,omitempty"`
}

type FindBenefitResponse struct {
	Benefits []string `json:"benefits,omitempty"`
}

// SearchQueuesRequest mirrors the subset of GET /queues parameters this app exposes. Benefit
// should be an exact name as returned by FindBenefitResponse -- an inexact one reliably returns
// zero results (NFZ's own dictionary search is fuzzy; the queues endpoint is not).
type SearchQueuesRequest struct {
	Province string `json:"province,omitempty"`
	Benefit  string `json:"benefit,omitempty"`
	Locality string `json:"locality,omitempty"`
	Case     Case   `json:"case,omitempty"`
	// ChildAgeBand, when set, filters to places serving that age bracket: one of
	// "0_3", "4_9", "10_15", "16_18". Maps to NFZ's benefitForChildren_* boolean params.
	ChildAgeBand string `json:"child_age_band,omitempty"`
	Page         int    `json:"page,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

// Queue is one wait-time forecast entry -- a specific provider + place + benefit combination.
// Deliberately flattened from NFZ's JSON:API-shaped {type,id,attributes{...}} envelope into a
// plain struct, since nothing in this app needs the JSON:API wrapper itself.
type Queue struct {
	Id       string `json:"id,omitempty"`
	Benefit  string `json:"benefit,omitempty"`
	Provider string `json:"provider,omitempty"`
	Place    string `json:"place,omitempty"`
	Address  string `json:"address,omitempty"`
	Locality string `json:"locality,omitempty"`
	Phone    string `json:"phone,omitempty"`

	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`

	// BenefitsForChildren is NFZ's own "Y"/"N" marker for whether this place treats children at
	// all -- surfaced as a plain bool so clients don't need to know NFZ's Y/N convention.
	BenefitsForChildren bool `json:"benefits_for_children,omitempty"`

	// EstimatedWait is NFZ's own free-text forecast, e.g. "15 dni" -- deliberately left as the
	// string NFZ publishes rather than parsed into a number, since its units/phrasing vary
	// ("do 30 dni", "1 miesiąc", ...) and NFZ is the authority on how to phrase it.
	EstimatedWait     string `json:"estimated_wait,omitempty"`
	WaitDataAsOf      string `json:"wait_data_as_of,omitempty"`
	AwaitingCount     int64  `json:"awaiting_count,omitempty"`
	AveragePeriodDays int64  `json:"average_period_days,omitempty"`

	Wheelchair bool `json:"wheelchair_accessible,omitempty"`
	ParkingLot bool `json:"parking_lot,omitempty"`
}

type SearchQueuesResponse struct {
	Queues []Queue `json:"queues,omitempty"`
	// Count is NFZ's own total match count (before pagination) -- lets the client show "12
	// placówek znaleziono" without fetching every page.
	Count int `json:"count,omitempty"`
}

// ProvinceForLocationResponse answers "which województwo is this coordinate in" -- used to
// preselect a sensible default province for the parent's NFZ search from their device location
// (or the Rybnik fallback), since the NFZ API itself has no lat/lon radius search.
type ProvinceForLocationResponse struct {
	Province Province `json:"province,omitempty"`
}
