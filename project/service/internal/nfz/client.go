package nfz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	nfzpg "github.com/Johncode248/Rodzic-Chat/project/common/nfz"
)

// baseURL is NFZ's own public "API Terminy Leczenia" v1.4 -- no API key, no auth, plain GET
// endpoints (see https://apinfz.nfz.gov.pl/app-itl-api-pcus/index.html). Kept as a constant
// rather than a variables.go secret since it isn't one -- it's a fixed public government API.
const baseURL = "https://apinfz.nfz.gov.pl/app-itl-api-pcus"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// NFZ publishes new wait-time data once a month, so successful answers are cached for a day: searches
// are instant, repeated ones don't hit NFZ's API (and its limits) at all.
const cacheTTL = 24 * time.Hour

type cacheEntry struct {
	body    []byte
	expires time.Time
}

var (
	cacheMu sync.Mutex
	cache   = map[string]cacheEntry{}
)

func cached(key string) ([]byte, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	e, ok := cache[key]
	if !ok || time.Now().After(e.expires) {
		delete(cache, key)
		return nil, false
	}
	return e.body, true
}

func store(key string, body []byte) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if len(cache) > 2000 { // crude bound: start over rather than grow without limit
		cache = map[string]cacheEntry{}
	}
	cache[key] = cacheEntry{body: body, expires: time.Now().Add(cacheTTL)}
}

func get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	reqURL := baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	if body, ok := cached(reqURL); ok {
		return body, nil
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

	store(reqURL, body)
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
// [req.Benefit] must be an exact name from FindBenefit (or nfz.Specialties); a partial/guessed one
// reliably returns zero rows (see this package's doc comment). With a position it runs the nearby search.
func SearchQueues(ctx context.Context, req nfzpg.SearchQueuesRequest) (nfzpg.SearchQueuesResponse, error) {
	if req.Latitude != 0 || req.Longitude != 0 {
		return searchNearby(ctx, req)
	}
	limit := req.Limit
	if limit <= 0 || limit > 25 {
		limit = 15
	}
	res, err := fetchQueues(ctx, req, req.Province, req.Page, limit)
	if err != nil {
		return res, err
	}
	for i := range res.Queues {
		res.Queues[i].DistanceKm = -1
	}
	return res, nil
}

// maxNearbyPages bounds the work of one nearby search: 25 results per page, per province.
const maxNearbyPages = 3

// searchNearby searches the two województwa whose capitals are nearest to the position (a search near a
// border must not stop at it), keeps what lies within the radius and sorts by distance or wait.
func searchNearby(ctx context.Context, req nfzpg.SearchQueuesRequest) (nfzpg.SearchQueuesResponse, error) {
	radius := float64(req.RadiusKm)
	if radius <= 0 {
		radius = 30
	}
	codes := nearestProvinces(req.Latitude, req.Longitude, 2)

	type result struct {
		queues []nfzpg.Queue
		err    error
	}
	results := make([]result, len(codes))
	var wg sync.WaitGroup
	for i, code := range codes {
		wg.Add(1)
		go func(i int, code string) {
			defer wg.Done()
			var all []nfzpg.Queue
			for page := 1; page <= maxNearbyPages; page++ {
				r, err := fetchQueues(ctx, req, code, page, 25)
				if err != nil {
					results[i] = result{all, err}
					return
				}
				all = append(all, r.Queues...)
				if len(r.Queues) < 25 {
					break
				}
			}
			results[i] = result{queues: all}
		}(i, code)
	}
	wg.Wait()

	seen := map[string]bool{}
	var out []nfzpg.Queue
	var firstErr error
	for _, r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		for _, q := range r.queues {
			if seen[q.Id] {
				continue
			}
			seen[q.Id] = true
			if q.Latitude == 0 && q.Longitude == 0 {
				continue // no position -> cannot tell whether it is nearby
			}
			q.DistanceKm = math.Round(haversineKm(req.Latitude, req.Longitude, q.Latitude, q.Longitude)*10) / 10
			if q.DistanceKm <= radius {
				out = append(out, q)
			}
		}
	}
	if len(out) == 0 && firstErr != nil {
		return nfzpg.SearchQueuesResponse{}, firstErr
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if req.Sort == "wait" {
			switch {
			case a.WaitDays >= 0 && b.WaitDays < 0:
				return true
			case a.WaitDays < 0 && b.WaitDays >= 0:
				return false
			case a.WaitDays != b.WaitDays:
				return a.WaitDays < b.WaitDays
			}
		}
		return a.DistanceKm < b.DistanceKm
	})
	count := len(out)
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := req.Page * limit // Page is 0-based here: the client asks for more by raising it
	if offset >= len(out) {
		out = nil
	} else {
		out = out[offset:]
		if len(out) > limit {
			out = out[:limit]
		}
	}
	return nfzpg.SearchQueuesResponse{Queues: out, Count: count}, nil
}

func nearestProvinces(lat, lon float64, n int) []string {
	type pd struct {
		code string
		d    float64
	}
	var list []pd
	for _, p := range provinces {
		list = append(list, pd{p.Code, haversineKm(lat, lon, p.Latitude, p.Longitude)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].d < list[j].d })
	var codes []string
	for i := 0; i < n && i < len(list); i++ {
		codes = append(codes, list[i].code)
	}
	return codes
}

// fetchQueues is one /queues call for one province and page.
func fetchQueues(ctx context.Context, req nfzpg.SearchQueuesRequest, province string, page, limit int) (nfzpg.SearchQueuesResponse, error) {
	query := url.Values{}
	if province != "" {
		query.Set("province", province)
	}
	if req.Benefit != "" {
		query.Set("benefit", req.Benefit)
	}
	if req.Locality != "" && req.Latitude == 0 && req.Longitude == 0 {
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
	query.Set("limit", strconv.Itoa(limit))
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
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
		q := nfzpg.Queue{
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
			DistanceKm:          -1,
		}
		q.WaitDays = waitDays(q)
		res.Queues = append(res.Queues, q)
	}
	return res, nil
}

// waitDays turns NFZ's data into one comparable number: its own forecast text ("15 dni", "2 miesiące",
// "3 tygodnie") when published, else the provider's average waiting period; -1 when NFZ has neither.
func waitDays(q nfzpg.Queue) int64 {
	if n, unit := leadingNumber(q.EstimatedWait); n > 0 {
		switch {
		case strings.HasPrefix(unit, "mies"):
			return n * 30
		case strings.HasPrefix(unit, "tyg"):
			return n * 7
		case strings.HasPrefix(unit, "rok"), strings.HasPrefix(unit, "lat"):
			return n * 365
		default:
			return n
		}
	}
	if q.AveragePeriodDays > 0 {
		return q.AveragePeriodDays
	}
	return -1
}

func leadingNumber(s string) (int64, string) {
	s = strings.TrimSpace(strings.ToLower(s))
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, ""
	}
	n, _ := strconv.ParseInt(s[:i], 10, 64)
	return n, strings.TrimSpace(s[i:])
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
