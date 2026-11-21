package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestFrozenFormatZeroDictionary(t *testing.T) {
	data, err := json.Marshal(Dictionary())
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	if hex.EncodeToString(hash[:]) != "744533032e5282766388956d37298cd41f2be0c6ba9e3a999cc897d03db0bed3" {
		t.Fatal("format-0 dictionary changed")
	}
}

func TestProviderRegistryAndPlanCoverage(t *testing.T) {
	known := map[uint32]string{0: "Self-managed", 1: "DigitalOcean", 2: "Akamai Cloud / Linode", 3: "Vultr", 4: "AWS", 65535: "Other"}
	providers := map[uint32]bool{}
	for _, p := range Providers {
		if providers[p.ID] || p.ID > 65535 || p.Name == "" {
			t.Fatalf("invalid provider: %+v", p)
		}
		providers[p.ID] = true
		if name, ok := known[p.ID]; ok && name != p.Name {
			t.Fatalf("existing provider ID changed: %d", p.ID)
		}
	}
	for id := range known {
		if !providers[id] {
			t.Fatalf("missing reserved provider %d", id)
		}
	}
	var plans []struct {
		Provider     uint32 `json:"provider"`
		Name         string `json:"name"`
		CPU          int    `json:"cpu"`
		RAM          int    `json:"ram"`
		Storage      *int   `json:"storage"`
		Architecture int    `json:"architecture"`
		Source       string `json:"source"`
	}
	if err := json.Unmarshal(Plans(), &plans); err != nil {
		t.Fatal(err)
	}
	coverage := map[uint32]int{}
	separateDisk := 0
	arm := 0
	for _, p := range plans {
		if !providers[p.Provider] || p.Name == "" || p.CPU < 0 || p.CPU >= len(CPU) || p.RAM < 0 || p.RAM >= len(RAM) || p.Source == "" {
			t.Fatalf("invalid plan: %+v", p)
		}
		if p.Storage != nil {
			if *p.Storage < 0 || *p.Storage >= len(Storage) {
				t.Fatal("invalid storage tier")
			}
		} else {
			separateDisk++
		}
		if p.Architecture == 1 {
			arm++
		}
		coverage[p.Provider]++
	}
	// Directory membership is independent of reviewed plan coverage.
	for _, id := range []uint32{0, 1, 2, 3, 4, 17, 18} {
		if coverage[id] == 0 {
			t.Fatalf("missing reviewed recommendations for %d", id)
		}
	}
	if separateDisk == 0 || arm == 0 {
		t.Fatal("separate-disk and Arm plans must remain represented")
	}
}
