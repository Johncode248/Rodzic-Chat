// Package rspopg holds the wire types for this app's integration with RSPO -- "Rejestr Szkół i
// Placówek Oświatowych" (the Ministry of Education's public register of every school and
// education-adjacent institution in Poland, https://rspo.gov.pl). Used so "Przedszkola", "Szkoły
// podstawowe i średnie", "Poradnie psychologiczno-pedagogiczne" (also the closest real registry
// for "Wczesne wspomaganie rozwoju", which those poradnie administer) and specialized
// rehabilitation centres (the closest real registry for "Terapia SI"/"ABA, spektrum autyzmu") in
// the Wyszukaj hub search real government data instead of a guessed OpenStreetMap keyword.
//
// NOT covered by RSPO at all (confirmed against its own institution-type dictionary): żłobki and
// kluby dziecięce -- those fall under a separate Ministry of Family registry with no usable
// public search API found, so that hub category still falls back to OpenStreetMap. Likewise
// korepetytorzy, szkoły językowe, nianie, kolonie/półkolonie -- inherently private/informal,
// no government registry exists for them.
package rspopg

// Category is a coarse grouping the Wyszukaj hub's education/support rows map to -- each expands
// to a handful of RSPO institution-type ids server-side (see rspo.CategoryTypeIds on the Go
// side).
type Category string

const (
	CategoryPreschool     Category = "przedszkole"
	CategorySchool        Category = "szkola"
	CategoryCounseling    Category = "poradnia"
	CategorySpecialCenter Category = "osrodek_specjalny"
)

type SearchResponse struct {
	Institutions []Institution `json:"institutions,omitempty"`
}

// Institution is one RSPO entry, flattened from its own nested JSON shape into what this app
// actually shows.
type Institution struct {
	Rspo       int    `json:"rspo,omitempty"`
	Name       string `json:"name,omitempty"`
	TypeName   string `json:"type_name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Website    string `json:"website,omitempty"`
	Email      string `json:"email,omitempty"`
	Address    string `json:"address,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Locality   string `json:"locality,omitempty"`
	StudentsNr int    `json:"students_nr,omitempty"`
}
