package geography

import "testing"

func TestIPCityCandidates(t *testing.T) {
	rows, err := MatchCity("US", "New York", "New York")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, city := range rows {
		if city.ID == 5128581 {
			found = true
		}
		if len(city.SearchNames) > 0 {
			t.Fatal("aliases exposed")
		}
	}
	if !found {
		t.Fatal("expected exact city candidate")
	}
	rows, err = MatchCity("US", "Missing place xyz", "")
	if err != nil || len(rows) != 0 {
		t.Fatal("invented candidate")
	}
	rows, err = MatchCity("XX", "New York", "")
	if err != nil || len(rows) != 0 {
		t.Fatal("unknown country candidate")
	}
}

func TestLookupIncludesHiddenPlaces(t *testing.T) {
	place, err := Lookup(5128581)
	if err != nil || !place.Found || place.Country.Code != "US" || place.City.ID != 5128581 {
		t.Fatal("New York lookup failed", err)
	}
	rows, err := readCountry("US")
	if err != nil {
		t.Fatal(err)
	}
	for _, city := range rows {
		if detailedFeature(city.Feature) {
			place, err := Lookup(city.ID)
			if err != nil || !place.Found || place.City.ID != city.ID {
				t.Fatal("hidden place cannot be resolved", err)
			}
			if len(place.City.SearchNames) > 0 {
				t.Fatal("search aliases exposed")
			}
			return
		}
	}
	t.Fatal("missing hidden-place fixture")
}
func TestMissingLookup(t *testing.T) {
	place, err := Lookup(4294967295)
	if err != nil || place.Found {
		t.Fatal("missing ID lookup")
	}
}

func TestCountriesAndCities(t *testing.T) {
	countries := Countries()
	if len(countries) < 200 {
		t.Fatal("country snapshot missing")
	}
	for _, country := range countries {
		result, err := Cities(country.Code, "", 0, true)
		if err != nil {
			t.Fatal(country.Code, err)
		}
		if result.Total != country.Cities || len(result.Cities) > 100 {
			t.Fatal("count or page mismatch", country.Code)
		}
		for _, city := range result.Cities {
			if city.ID == 0 || city.Name == "" {
				t.Fatal("invalid city")
			}
		}
	}
	result, err := Cities("TW", "Taipei", 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range result.Cities {
		if c.ID == 1668341 {
			found = true
		}
	}
	if !found {
		t.Fatal("Taipei missing")
	}
	us, err := Cities("US", "", 0)
	if err != nil || !us.More || len(us.Cities) != 100 {
		t.Fatal("pagination missing", err)
	}
	next, err := Cities("US", "", 100)
	if err != nil || next.Cities[0].ID == us.Cities[0].ID {
		t.Fatal("pagination duplicated", err)
	}
	empty, err := Cities("TW", "unlikely-city-name-xyz", 0)
	if err != nil || empty.Total != 0 || empty.Cities == nil {
		t.Fatal("empty result")
	}
	for _, code := range []string{"../", "XX", "", "USA"} {
		if _, err := Cities(code, "", 0); err == nil {
			t.Fatal("invalid country accepted", code)
		}
	}
}

func TestDefaultFilterAndRanking(t *testing.T) {
	simple, err := Cities("US", "New York", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(simple.Cities) == 0 || simple.Cities[0].ID != 5128581 {
		t.Fatal("exact city name did not rank first")
	}
	for _, city := range simple.Cities {
		if detailedFeature(city.Feature) {
			t.Fatal("detailed feature in default results")
		}
	}
	all, err := Cities("US", "", 0, true)
	if err != nil {
		t.Fatal(err)
	}
	basic, err := Cities("US", "", 0)
	if err != nil || all.Total <= basic.Total {
		t.Fatal("details filter ineffective")
	}
	exact := City{Name: "Example City"}
	admin := City{Name: "Other", Admin: "Example City"}
	partial := City{Name: "Example City East"}
	if searchRank(exact, "example city") >= searchRank(partial, "example city") || searchRank(partial, "example city") >= searchRank(admin, "example city") {
		t.Fatal("ranking order")
	}
	for _, code := range []string{"PPL", "PPLC", "PPLA", "PPLA2", "PPLA3", "PPLA4", "PPLA5"} {
		if detailedFeature(code) {
			t.Fatal("ordinary settlement removed")
		}
	}
}

func TestLocalizedNamesAndSearch(t *testing.T) {
	examples := []struct {
		country, query string
		id             uint32
	}{{"TW", "\u81fa\u5317", 1668341}, {"JP", "\u6771\u4eac", 1850147}, {"US", "\u7d10\u7d04", 5128581}, {"DE", "\u67cf\u6797", 2950159}, {"FR", "\u5df4\u9ece", 2988507}}
	for _, x := range examples {
		result, err := Cities(x.country, x.query, 0)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, c := range result.Cities {
			if len(c.SearchNames) != 0 {
				t.Fatal("internal search aliases exposed")
			}
			if c.ID == x.id {
				found = true
				if c.Names["zh-TW"] == "" || c.Names["en"] == "" {
					t.Fatal("translation missing")
				}
			}
		}
		if !found {
			t.Fatalf("localized search failed for %s", x.country)
		}
	}
	result, err := Cities("TW", "Taipei", 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range result.Cities {
		if c.ID == 1668341 {
			found = true
		}
	}
	if !found {
		t.Fatal("English alias lost")
	}
}
