package nfz

import "math"

// Provinces is NFZ's own fixed 16-entry code list (see the `province` parameter description on
// every /benefits, /providers, /places, /localities, /streets and /queues endpoint) paired with
// each województwo's capital city coordinates, used only as an approximate reverse-geocode
// anchor by ProvinceForLocation below -- picking the nearest capital by straight-line distance
// is good enough to preselect a sensible default province for a search UI; it deliberately
// doesn't need to be exact since the parent can always override it manually.
var provinces = []struct {
	Code      string
	Name      string
	Latitude  float64
	Longitude float64
}{
	{"01", "dolnośląskie", 51.1079, 17.0385},
	{"02", "kujawsko-pomorskie", 53.1235, 18.0084},
	{"03", "lubelskie", 51.2465, 22.5684},
	{"04", "lubuskie", 52.7368, 15.2288},
	{"05", "łódzkie", 51.7592, 19.4560},
	{"06", "małopolskie", 50.0647, 19.9450},
	{"07", "mazowieckie", 52.2297, 21.0122},
	{"08", "opolskie", 50.6751, 17.9213},
	{"09", "podkarpackie", 50.0413, 21.9990},
	{"10", "podlaskie", 53.1325, 23.1688},
	{"11", "pomorskie", 54.3520, 18.6466},
	{"12", "śląskie", 50.2649, 19.0238},
	{"13", "świętokrzyskie", 50.8661, 20.6286},
	{"14", "warmińsko-mazurskie", 53.7784, 20.4801},
	{"15", "wielkopolskie", 52.4064, 16.9252},
	{"16", "zachodniopomorskie", 53.4285, 14.5528},
}

// ProvinceForLocation returns NFZ's own two-digit code and Polish name for the województwo
// whose capital is nearest to (lat, lon) -- see the doc comment on [provinces] for why "nearest
// capital" is an acceptable approximation here.
func ProvinceForLocation(lat, lon float64) (code string, name string) {
	best := provinces[0]
	bestDist := haversineKm(lat, lon, best.Latitude, best.Longitude)
	for _, p := range provinces[1:] {
		d := haversineKm(lat, lon, p.Latitude, p.Longitude)
		if d < bestDist {
			bestDist = d
			best = p
		}
	}
	return best.Code, best.Name
}

// ProvinceName returns NFZ's Polish name for a two-digit province code, or "" if unknown.
func ProvinceName(code string) string {
	for _, p := range provinces {
		if p.Code == code {
			return p.Name
		}
	}
	return ""
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }

	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}
