package nfz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
)

// baseURL is NFZ's own public "API Terminy Leczenia" v1.4 -- no API key, no auth, plain GET
// endpoints (see https://apinfz.nfz.gov.pl/app-itl-api-pcus/index.html). Kept as a constant
// rather than a variables.go secret since it isn't one -- it's a fixed public government API.
const baseURL = "https://apinfz.nfz.gov.pl/app-itl-api-pcus"

var httpClient = &http.Client{Timeout: 10 * time.Second}

func get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	reqURL := baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("nfz: %s returned status %d: %s", path, resp.StatusCode, string(body))
	}

	return body, nil
}

// dictionaryResponse mirrors NFZ's shared envelope for /benefits, /providers, /places,
// /localities, /streets -- {"data": [...names]}. See "benefits-response" etc. in the v1.4
// schema; they're all shaped the same way (a flat array of matching strings).
type dictionaryResponse struct {
	Data []string `json:"data"`
}

// FindBenefit resolves free text into exact NFZ benefit names via GET /benefits -- NFZ's own
// search there is a fuzzy/stemmed match (confirmed empirically: "dziecięc" matches
// "ORTOPEDII..." entries with no substring in common), unlike /queues' benefit filter, which
// needs an exact name. Always call this before SearchQueues.
func FindBenefit(ctx context.Context, query string) ([]string, error) {
	body, err := get(ctx, "/benefits", url.Values{"name": {query}, "limit": {"25"}})
	if err != nil {
		return nil, err
	}
	var parsed dictionaryResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("nfz: unexpected /benefits response: %w", err)
	}
	return parsed.Data, nil
}

type nfzProviderData struct {
	Awaiting      int64 `json:"awaiting"`
	AveragePeriod int64 `json:"average-period"`
}

// nfzQueuesResponse mirrors GET /queues' JSON:API-shaped body -- see the "queue"/"queue-
// attributes" schemas in the v1.4 spec. Only the fields this app surfaces are declared; NFZ
// returns many more (accessibility flags, building type, etc.) that aren't used here.
type nfzQueuesResponse struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
	Data []struct {
		Id         string `json:"id"`
		Attributes struct {
			Benefit             string  `json:"benefit"`
			Provider            string  `json:"provider"`
			Place               string  `json:"place"`
			Address             string  `json:"address"`
			Locality            string  `json:"locality"`
			Phone               string  `json:"phone"`
			Latitude            float64 `json:"latitude"`
			Longitude           float64 `json:"longitude"`
			BenefitsForChildren string  `json:"benefits-for-children"`
			Wheelchairs         string  `json:"wheelchairs"`
			CarPark             string  `json:"car-park"`
			Statistics          struct {
				ProviderData *nfzProviderData `json:"provider-data"`
			} `json:"statistics"`
			Dates struct {
				Applicable        bool   `json:"applicable"`
				Pcus              string `json:"pcus"`
				DateSituationAsAt string `json:"date-situation-as-at"`
			} `json:"dates"`
		} `json:"attributes"`
	} `json:"data"`
}

// SearchQueues calls GET /queues -- the actual wait-time forecast + provider directory search.
// [req.Benefit] must be an exact name from FindBenefit; a partial/guessed one reliably returns
// zero rows (see this package's doc comment).
func SearchQueues(ctx context.Context, req nfzpg.SearchQueuesRequest) (nfzpg.SearchQueuesResponse, error) {
	query := url.Values{}
	if req.Province != "" {
		query.Set("province", req.Province)
	}
	if req.Benefit != "" {
		query.Set("benefit", req.Benefit)
	}
	if req.Locality != "" {
		query.Set("locality", req.Locality)
	}
	if req.Case == nfzpg.CaseUrgent {
		query.Set("case", "2")
	}
	switch req.ChildAgeBand {
	case "0_3":
		query.Set("benefitForChildren_0_3", "true")
	case "4_9":
		query.Set("benefitForChildren_4_9", "true")
	case "10_15":
		query.Set("benefitForChildren_10_15", "true")
	case "16_18":
		query.Set("benefitForChildren_16_18", "true")
	}
	limit := req.Limit
	if limit <= 0 || limit > 25 {
		limit = 15
	}
	query.Set("limit", strconv.Itoa(limit))
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}

	body, err := get(ctx, "/queues", query)
	if err != nil {
		return nfzpg.SearchQueuesResponse{}, err
	}

	var parsed nfzQueuesResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nfzpg.SearchQueuesResponse{}, fmt.Errorf("nfz: unexpected /queues response: %w", err)
	}

	res := nfzpg.SearchQueuesResponse{Count: parsed.Meta.Count}
	for _, item := range parsed.Data {
		a := item.Attributes
		res.Queues = append(res.Queues, nfzpg.Queue{
			Id:                  item.Id,
			Benefit:             a.Benefit,
			Provider:            a.Provider,
			Place:               a.Place,
			Address:             a.Address,
			Locality:            a.Locality,
			Phone:               a.Phone,
			Latitude:            a.Latitude,
			Longitude:           a.Longitude,
			BenefitsForChildren: a.BenefitsForChildren == "Y",
			EstimatedWait:       a.Dates.Pcus,
			WaitDataAsOf:        a.Dates.DateSituationAsAt,
			AwaitingCount:       providerAwaiting(a.Statistics.ProviderData),
			AveragePeriodDays:   providerAveragePeriod(a.Statistics.ProviderData),
			Wheelchair:          a.Wheelchairs == "Y",
			ParkingLot:          a.CarPark == "Y",
		})
	}
	return res, nil
}

func providerAwaiting(d *nfzProviderData) int64 {
	if d == nil {
		return 0
	}
	return d.Awaiting
}

func providerAveragePeriod(d *nfzProviderData) int64 {
	if d == nil {
		return 0
	}
	return d.AveragePeriod
}
