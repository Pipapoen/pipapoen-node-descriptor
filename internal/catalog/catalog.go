package catalog

import (
	"embed"
	"encoding/json"
)

//go:embed dictionary.json plans.json
var files embed.FS

func Dictionary() [][3]uint32 {
	b, _ := files.ReadFile("dictionary.json")
	var v [][3]uint32
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}
	return v
}
func Plans() json.RawMessage { b, _ := files.ReadFile("plans.json"); return b }

var Providers = []struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}{
	{0, "Self-managed"}, {1, "DigitalOcean"}, {2, "Akamai Cloud / Linode"}, {3, "Vultr"}, {4, "AWS"}, {65535, "Other"},
}
var CPU = []string{"(0, 1]", "(1, 2]", "(2, 3]", "(3, 4]", "(4, 6]", "(6, 8]", "(8, 12]", "(12, 16]", "(16, 24]", "(24, 32]", "(32, 48]", "(48, 64]", "(64, 96]", "(96, 128]", "(128, 256]", "> 256"}
var RAM = []string{"(0, 0.5]", "(0.5, 1]", "(1, 2]", "(2, 3]", "(3, 4]", "(4, 6]", "(6, 8]", "(8, 12]", "(12, 16]", "(16, 24]", "(24, 32]", "(32, 64]", "(64, 128]", "(128, 256]", "(256, 512]", "> 512"}
var Storage = []string{"(0, 32]", "(32, 64]", "(64, 128]", "(128, 256]", "(256, 512]", "(512, 1024]", "(1024, 2048]", "> 2048"}
