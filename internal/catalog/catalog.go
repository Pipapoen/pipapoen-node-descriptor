package catalog

import (
	"embed"
	"encoding/json"
)

//go:embed dictionary.json plans.json providers.json catalog-info.json
var files embed.FS

func Dictionary() [][3]uint32 {
	b, _ := files.ReadFile("dictionary.json")
	var v [][3]uint32
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}
	return v
}

var plans = readJSON("plans.json")
var info = readJSON("catalog-info.json")

func readJSON(name string) json.RawMessage {
	b, err := files.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return b
}
func Plans() json.RawMessage { return plans }
func Info() json.RawMessage  { return info }

type Provider struct {
	ID        uint32   `json:"id"`
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases"`
	Website   string   `json:"website,omitempty"`
	CheckedAt string   `json:"checkedAt,omitempty"`
}

var Providers = func() []Provider {
	var providers []Provider
	if err := json.Unmarshal(readJSON("providers.json"), &providers); err != nil {
		panic(err)
	}
	return providers
}()
var CPU = []string{"(0, 1]", "(1, 2]", "(2, 3]", "(3, 4]", "(4, 6]", "(6, 8]", "(8, 12]", "(12, 16]", "(16, 24]", "(24, 32]", "(32, 48]", "(48, 64]", "(64, 96]", "(96, 128]", "(128, 256]", "> 256"}
var RAM = []string{"(0, 0.5]", "(0.5, 1]", "(1, 2]", "(2, 3]", "(3, 4]", "(4, 6]", "(6, 8]", "(8, 12]", "(12, 16]", "(16, 24]", "(24, 32]", "(32, 64]", "(64, 128]", "(128, 256]", "(256, 512]", "> 512"}
var Storage = []string{"(0, 32]", "(32, 64]", "(64, 128]", "(128, 256]", "(256, 512]", "(512, 1024]", "(1024, 2048]", "> 2048"}
