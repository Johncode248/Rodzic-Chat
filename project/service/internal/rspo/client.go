package rspo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	rspopg "github.com/Johncode248/Rodzic-Chat/project/common/rspo"
)

// baseURL is RSPO's own public search API -- reverse-engineered from the network calls its own
// search UI (https://rspo.gov.pl) makes; no official public API docs or key exist for it, it's a
// fixed public government endpoint like the NFZ one (see project/service/internal/nfz).
const baseURL = "https://rspo.gov.pl/api"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// categoryTypeIds maps each Wyszukaj hub category to the RSPO institution-type ids (see
// GET /api/Dictionary/InstitutionType) that belong to it. Filtering happens client-side over the
// name/locality-matched page RSPO returns (see Search below) -- RSPO's own API has no combined
// type+text filter, so this is the practical way to combine "matches this text" with "is this
// kind of institution" in one call.
var categoryTypeIds = map[rspopg.Category][]int{
	rspopg.CategoryPreschool: {1, 81, 80}, // Przedszkole, Punkt przedszkolny, Zespół wychowania przedszkolnego
	rspopg.CategorySchool: {
		3,   // Szkoła podstawowa
		14,  // Liceum ogólnokształcące
		16,  // Technikum
		93,  // Branżowa szkoła I stopnia
		94,  // Branżowa szkoła II stopnia
		19,  // Szkoła policealna
		100, // Zespół szkół i placówek oświatowych
	},
	rspopg.CategoryCounseling: {
		48, // Poradnia psychologiczno-pedagogiczna -- also administers Wczesne wspomaganie rozwoju (WWR)
		49, // Poradnia specjalistyczna
	},
	rspopg.CategorySpecialCenter: {
		52, // Ośrodek Rewalidacyjno-Wychowawczy -- closest RSPO match for terapia SI/ABA, spektrum autyzmu
		51, // Specjalny Ośrodek Szkolno-Wychowawczy
		50, // Specjalny Ośrodek Wychowawczy
	},
}

type institutionLocality struct {
	Name string `json:"name"`
}

type institutionType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type rspoInstitution struct {
	Rspo                int                 `json:"rspo"`
	Name                string              `json:"name"`
	Type                institutionType     `json:"type"`
	Telephone           string              `json:"telephone"`
	Website             string              `json:"website"`
	Email               string              `json:"email"`
	HqAddressZipCode    string              `json:"hqAddressZipCode"`
	HqAddressBuildingNr string              `json:"hqAddressBuildingNr"`
	HqAddressPremiseNr  string              `json:"hqAddressPremiseNr"`
	HqAddressStreet     string              `json:"hqAddressStreet"`
	HqAddressLocality   institutionLocality `json:"hqAddressLocality"`
	StudentsNr          int                 `json:"studentsNr"`
}

type rspoSearchResponse struct {
	TotalCount int               `json:"totalCount"`
	Items      []rspoInstitution `json:"items"`
}

// Search calls RSPO's institution search with free text [query] (matches institution name --
// Polish institution names conventionally include their city, so a city name alone works
// reasonably well too) and keeps only results whose type belongs to [category], up to [limit].
// A blank query still works: it falls back to the category's own Polish label so results stay
// on-topic instead of returning an arbitrary slice of all ~56k RSPO institutions.
func Search(ctx context.Context, query string, category rspopg.Category, limit int) ([]rspopg.Institution, error) {
	typeIds := categoryTypeIds[category]

	effectiveQuery := strings.TrimSpace(query)
	if effectiveQuery == "" {
		effectiveQuery = categoryFallbackQuery(category)
	}

	q := url.Values{}
	q.Set("PageOffset", "0")
	// Fetches a wider page than [limit] since results are filtered client-side by type
	// afterward -- a plain name/city match can return institutions of unrelated types mixed
	// together (e.g. searching a city returns every school type in it).
	q.Set("PageSize", "80")
	q.Set("IsDescending", "false")
	q.Set("OrderBy", "")
	q.Set("Filter", effectiveQuery)

	reqURL := baseURL + "/Institution?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rspo: /Institution returned status %d: %s", resp.StatusCode, string(body))
	}

	var parsed rspoSearchResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("rspo: unexpected /Institution response: %w", err)
	}

	allowedType := func(typeId int) bool {
		if len(typeIds) == 0 {
			return true
		}
		for _, id := range typeIds {
			if id == typeId {
				return true
			}
		}
		return false
	}

	res := make([]rspopg.Institution, 0, limit)
	for _, item := range parsed.Items {
		if !allowedType(item.Type.Id) {
			continue
		}
		res = append(res, rspopg.Institution{
			Rspo:       item.Rspo,
			Name:       item.Name,
			TypeName:   item.Type.Name,
			Phone:      item.Telephone,
			Website:    item.Website,
			Email:      item.Email,
			Address:    formatAddress(item.HqAddressStreet, item.HqAddressBuildingNr, item.HqAddressPremiseNr),
			PostalCode: item.HqAddressZipCode,
			Locality:   item.HqAddressLocality.Name,
			StudentsNr: item.StudentsNr,
		})
		if len(res) >= limit {
			break
		}
	}

	return res, nil
}

func formatAddress(street, buildingNr, premiseNr string) string {
	address := strings.TrimSpace(street)
	if buildingNr != "" {
		address = strings.TrimSpace(address + " " + buildingNr)
	}
	if premiseNr != "" {
		address += "/" + premiseNr
	}
	return address
}

func categoryFallbackQuery(category rspopg.Category) string {
	switch category {
	case rspopg.CategoryPreschool:
		return "przedszkole"
	case rspopg.CategorySchool:
		return "szkoła"
	case rspopg.CategoryCounseling:
		return "poradnia"
	case rspopg.CategorySpecialCenter:
		return "ośrodek"
	default:
		return ""
	}
}
