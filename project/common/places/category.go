package placespg

// Category is a flat key (e.g. "pediatra") grouped for display purposes
// under a CategoryGroup (e.g. "zdrowie"). Stored as plain text on Place
// rather than a DB enum -- validated against AllCategoryKeys() at the
// application layer instead, so adding a category never needs a migration.
type Category struct {
	Key   string `json:"key,omitempty"`
	Label string `json:"label,omitempty"`
}

type CategoryGroup struct {
	Key        string     `json:"key,omitempty"`
	Label      string     `json:"label,omitempty"`
	Categories []Category `json:"categories,omitempty"`
}

// CategoryGroups mirrors the category list a parent asked for when
// specifying this feature -- grouped the same way for the picker UI.
func CategoryGroups() []CategoryGroup {
	return []CategoryGroup{
		{
			Key: "zdrowie", Label: "Zdrowie i opieka medyczna",
			Categories: []Category{
				{Key: "pediatra", Label: "Pediatra / lekarz rodzinny"},
				{Key: "alergolog_dzieciecy", Label: "Alergolog dziecięcy"},
				{Key: "dermatolog_dzieciecy", Label: "Dermatolog dziecięcy"},
				{Key: "ortopeda_dzieciecy", Label: "Ortopeda dziecięcy"},
				{Key: "okulista_dzieciecy", Label: "Okulista dziecięcy"},
				{Key: "laryngolog_dzieciecy", Label: "Laryngolog dziecięcy"},
				{Key: "endokrynolog_dzieciecy", Label: "Endokrynolog dziecięcy"},
				{Key: "stomatolog_dzieciecy", Label: "Stomatolog dziecięcy / ortodonta"},
				{Key: "logopeda", Label: "Logopeda"},
				{Key: "psycholog_dzieciecy", Label: "Psycholog / psychiatra dziecięcy"},
				{Key: "fizjoterapeuta_dziecieca", Label: "Fizjoterapeuta dziecięcy"},
				{Key: "poradnia_laktacyjna", Label: "Poradnia laktacyjna / położna"},
				{Key: "szpital_pediatryczny", Label: "Szpital / SOR pediatryczny"},
				{Key: "apteka_calodobowa", Label: "Apteka (całodobowa/dyżurna)"},
				{Key: "punkt_szczepien", Label: "Punkt szczepień"},
				{Key: "laboratorium", Label: "Laboratorium / USG / RTG"},
			},
		},
		{
			Key: "edukacja", Label: "Edukacja i opieka",
			Categories: []Category{
				{Key: "zlobek", Label: "Żłobek / klub dziecięcy"},
				{Key: "przedszkole", Label: "Przedszkole"},
				{Key: "szkola_podstawowa", Label: "Szkoła podstawowa"},
				{Key: "szkola_srednia", Label: "Szkoła średnia"},
				{Key: "korepetycje", Label: "Korepetytorzy"},
				{Key: "szkola_jezykowa", Label: "Szkoła językowa"},
				{Key: "niania", Label: "Niania / opiekunka"},
				{Key: "kolonie", Label: "Kolonie i obozy"},
				{Key: "polkolonie", Label: "Półkolonie"},
			},
		},
		{
			Key: "zajecia", Label: "Zajęcia pozalekcyjne i sport",
			Categories: []Category{
				{Key: "klub_sportowy", Label: "Klub sportowy"},
				{Key: "szkolka_plywacka", Label: "Szkółka pływacka"},
				{Key: "zajecia_artystyczne", Label: "Zajęcia artystyczne / muzyczne"},
				{Key: "zajecia_rozwojowe", Label: "Zajęcia rozwojowe (szachy, robotyka)"},
				{Key: "harcerstwo", Label: "Harcerstwo / zuchy"},
			},
		},
		{
			Key: "rozwoj", Label: "Rozwój i wsparcie specjalistyczne",
			Categories: []Category{
				{Key: "wwr", Label: "Wczesne wspomaganie rozwoju"},
				{Key: "terapia_si", Label: "Terapia integracji sensorycznej"},
				{Key: "terapia_aba", Label: "Terapia behawioralna (ABA)"},
				{Key: "poradnia_psychologiczno_pedagogiczna", Label: "Poradnia psychologiczno-pedagogiczna"},
				{Key: "diagnostyka_dysleksji", Label: "Diagnostyka dysleksji/dysgrafii"},
			},
		},
		{
			Key: "wolny_czas", Label: "Czas wolny i rozrywka",
			Categories: []Category{
				{Key: "plac_zabaw", Label: "Plac zabaw / sala zabaw"},
				{Key: "park_trampolin_baseny", Label: "Park trampolin / basen / aquapark"},
				{Key: "kino_teatr", Label: "Kino / teatr dla dzieci"},
				{Key: "muzeum", Label: "Muzeum / wystawa interaktywna"},
				{Key: "urodziny_dla_dzieci", Label: "Urodziny dla dzieci"},
				{Key: "zoo_gospodarstwo", Label: "Zoo / gospodarstwo edukacyjne"},
				{Key: "biblioteka", Label: "Biblioteka / mediateka"},
			},
		},
		{
			Key: "zakupy", Label: "Zakupy i produkty",
			Categories: []Category{
				{Key: "odziez_dziecieca", Label: "Odzież dziecięca"},
				{Key: "zabawki", Label: "Zabawki"},
				{Key: "wypozyczalnia_sprzetu", Label: "Wypożyczalnia sprzętu dziecięcego"},
				{Key: "wyprawka_noworodka", Label: "Wyprawka dla noworodka"},
			},
		},
		{
			Key: "formalnosci", Label: "Formalności i wsparcie rodzinne",
			Categories: []Category{
				{Key: "urzad", Label: "Urząd (rejestracja, świadczenia)"},
				{Key: "poradnia_rodzinna", Label: "Poradnia rodzinna / szkoła dla rodziców"},
				{Key: "grupa_wsparcia", Label: "Grupa wsparcia dla rodziców"},
				{Key: "inne", Label: "Inne"},
			},
		},
	}
}

func IsValidCategory(key string) bool {
	for _, group := range CategoryGroups() {
		for _, category := range group.Categories {
			if category.Key == key {
				return true
			}
		}
	}
	return false
}

func CategoryLabel(key string) string {
	for _, group := range CategoryGroups() {
		for _, category := range group.Categories {
			if category.Key == key {
				return category.Label
			}
		}
	}
	return key
}
