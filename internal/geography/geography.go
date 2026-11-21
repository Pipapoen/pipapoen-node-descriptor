package geography

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Data is a versioned, attributed GeoNames snapshot. Each country is compressed independently.
//
//go:embed data/countries.json data/manifest.json data/index.json data/cities/*.json.gz
var data embed.FS
var indexOnce sync.Once
var countryIndex map[string]string

type Country struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Cities int    `json:"cities"`
}
type City struct {
	ID          uint32            `json:"id"`
	Name        string            `json:"name"`
	ASCIIName   string            `json:"asciiName"`
	Admin       string            `json:"admin"`
	Feature     string            `json:"feature"`
	Population  uint64            `json:"population"`
	Names       map[string]string `json:"names,omitempty"`
	AdminNames  map[string]string `json:"adminNames,omitempty"`
	SearchNames []string          `json:"searchNames,omitempty"`
}
type Result struct {
	Cities []City `json:"cities"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	More   bool   `json:"more"`
}

func Countries() []Country {
	b, _ := data.ReadFile("data/countries.json")
	var out []Country
	_ = json.Unmarshal(b, &out)
	return out
}

type Location struct {
	Found   bool    `json:"found"`
	Country Country `json:"country"`
	City    City    `json:"city"`
}

func Lookup(id uint32) (Location, error) {
	indexOnce.Do(func() { b, _ := data.ReadFile("data/index.json"); _ = json.Unmarshal(b, &countryIndex) })
	code, ok := countryIndex[strconv.FormatUint(uint64(id), 10)]
	if !ok {
		return Location{}, nil
	}
	rows, err := readCountry(code)
	if err != nil {
		return Location{}, err
	}
	for _, city := range rows {
		if city.ID == id {
			city.SearchNames = nil
			for _, country := range Countries() {
				if country.Code == code {
					return Location{true, country, city}, nil
				}
			}
		}
	}
	return Location{}, nil
}

// MatchCity returns explicit suggestions from matching country and city names.
// It never assigns an inferred identifier to the IP database record.
func MatchCity(country, name, admin string) ([]City, error) {
	if len(country) != 2 || strings.TrimSpace(name) == "" {
		return []City{}, nil
	}
	rows, err := readCountry(strings.ToUpper(country))
	if err != nil {
		return []City{}, nil
	}
	name = strings.ToLower(strings.TrimSpace(name))
	admin = strings.ToLower(strings.TrimSpace(admin))
	matches := []City{}
	for _, city := range rows {
		same := strings.ToLower(city.Name) == name || strings.ToLower(city.ASCIIName) == name
		for _, v := range city.Names {
			same = same || strings.ToLower(v) == name
		}
		if same {
			city.SearchNames = nil
			matches = append(matches, city)
		}
	}
	adminMatch := func(c City) bool {
		if admin == "" {
			return false
		}
		if strings.ToLower(c.Admin) == admin {
			return true
		}
		for _, v := range c.AdminNames {
			if strings.ToLower(v) == admin {
				return true
			}
		}
		return false
	}
	sort.SliceStable(matches, func(i, j int) bool { return adminMatch(matches[i]) && !adminMatch(matches[j]) })
	if len(matches) > 10 {
		matches = matches[:10]
	}
	return matches, nil
}
func readCountry(code string) ([]City, error) {
	b, err := data.ReadFile("data/cities/" + code + ".json.gz")
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	var rows []City
	err = json.NewDecoder(io.LimitReader(gz, 8*1024*1024)).Decode(&rows)
	return rows, err
}
func Cities(country, query string, offset int, details ...bool) (Result, error) {
	country = strings.ToUpper(country)
	if len(country) != 2 || country[0] < 'A' || country[0] > 'Z' || country[1] < 'A' || country[1] > 'Z' || offset < 0 || offset > 100000 || len(query) > 100 {
		return Result{}, errors.New("invalid geography request")
	}
	b, err := data.ReadFile("data/cities/" + country + ".json.gz")
	if err != nil {
		return Result{}, errors.New("unknown country")
	}
	gz, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return Result{}, err
	}
	defer gz.Close()
	var all []City
	if err := json.NewDecoder(io.LimitReader(gz, 8*1024*1024)).Decode(&all); err != nil {
		return Result{}, err
	}
	type ranked struct {
		city City
		rank int
	}
	matchedRows := make([]ranked, 0, len(all))
	includeDetails := len(details) > 0 && details[0]
	query = strings.ToLower(strings.TrimSpace(query))
	for _, c := range all {
		if !includeDetails && detailedFeature(c.Feature) {
			continue
		}
		rank := searchRank(c, query)
		if rank >= 0 {
			c.SearchNames = nil
			matchedRows = append(matchedRows, ranked{c, rank})
		}
	}
	sort.SliceStable(matchedRows, func(i, j int) bool { return matchedRows[i].rank < matchedRows[j].rank })
	matches := make([]City, 0, len(matchedRows))
	for _, row := range matchedRows {
		matches = append(matches, row.city)
	}
	start := min(offset, len(matches))
	end := min(start+100, len(matches))
	return Result{matches[start:end], len(matches), start, end < len(matches)}, nil
}

func detailedFeature(code string) bool {
	switch code {
	case "PPLX", "PPLH", "PPLQ", "PPLW":
		return true
	}
	return false
}
func searchRank(c City, query string) int {
	if query == "" {
		return 0
	}
	best := 4
	names := []string{c.Name, c.ASCIIName}
	for _, v := range c.Names {
		names = append(names, v)
	}
	for _, name := range names {
		name = strings.ToLower(name)
		if name == query {
			return 0
		}
		if strings.HasPrefix(name, query) {
			best = min(best, 1)
		} else if strings.Contains(name, query) {
			best = min(best, 2)
		}
	}
	if best < 4 {
		return best
	}
	if strings.Contains(strings.ToLower(c.Admin), query) {
		return 3
	}
	for _, name := range c.AdminNames {
		if strings.Contains(strings.ToLower(name), query) {
			return 3
		}
	}
	for _, name := range c.SearchNames {
		if strings.Contains(strings.ToLower(name), query) {
			return 3
		}
	}
	return -1
}
