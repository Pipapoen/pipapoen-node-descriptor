package codec

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/catalog"
)

func TestIndependentPackingVectors(t *testing.T) {
	e := engine(t)
	b, err := os.ReadFile("testdata/packing.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Descriptor Descriptor `json:"descriptor"`
		Payload    string     `json:"payload"`
		Bits       int        `json:"bits"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		payload, bits, _, err := e.pack(row.Descriptor)
		if err != nil || payload != row.Payload || bits != row.Bits {
			t.Fatal("independent packing vector mismatch", err)
		}
	}
}

func TestGeoIDDecodingDoesNotRequirePickerRecord(t *testing.T) {
	e := engine(t)
	for _, id := range []uint32{1, 4294967295} {
		d := Descriptor{NodeID: "0123456789abcdef0123", Geo: ptr(id)}
		out, err := e.Encode(d, "n", "example.org", false)
		if err != nil {
			t.Fatal(err)
		}
		got, err := e.Decode(out.Hostname)
		if err != nil || got.Geo == nil || *got.Geo != id {
			t.Fatal("numeric ID depends on reference lookup", err)
		}
	}
}

func engine(t *testing.T) *Engine {
	t.Helper()
	key := make([]byte, 32)
	e, err := New(key, catalog.Dictionary())
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func ptr(v uint32) *uint32 { return &v }
func TestAllFieldMasks(t *testing.T) {
	e := engine(t)
	for mask := 0; mask < 512; mask++ {
		for _, maximum := range []bool{false, true} {
			d := Descriptor{NodeID: strings.Repeat("0", 20)}
			if maximum {
				d.NodeID = strings.Repeat("f", 20)
			}
			p := []**uint32{&d.Geo, &d.Provider, &d.Architecture, &d.NodeType, &d.CPU, &d.RAM, &d.Storage}
			for i, target := range p {
				if mask&(1<<i) != 0 {
					v := uint32(0)
					if maximum {
						v = uint32(uint64(1)<<widths[i] - 1)
					}
					if i == 0 && v == 0 {
						v = 1
					}
					*target = ptr(v)
				}
			}
			if mask&128 != 0 {
				d.IPv4 = "192.0.2.10"
			}
			if mask&256 != 0 {
				d.IPv6 = "2001:db8::10"
			}
			for _, padded := range []bool{false, true} {
				out, err := e.Encode(d, "n", "example.org", padded)
				if err != nil {
					t.Fatalf("mask %d: %v", mask, err)
				}
				if len(out.Label) > 63 || padded && len(out.Label) != 63 {
					t.Fatal("length")
				}
				got, err := e.Decode("https://" + strings.ToUpper(out.Hostname) + "/path")
				if err != nil || !Same(d, got) {
					t.Fatalf("mask %d: %v %+v", mask, err, got)
				}
				if len(out.Label) != out.Preview.LabelLength {
					t.Fatal("preview mismatch")
				}
			}
		}
	}
}
func TestInvalidInputs(t *testing.T) {
	e := engine(t)
	for _, input := range []string{"", strings.Repeat("a", 64), "n9" + strings.Repeat("0", 30), "n0_" + strings.Repeat("0", 30), "file://example.org", "https://user:pass@example.org"} {
		if _, err := e.Decode(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	for _, d := range []Descriptor{{NodeID: "x"}, {NodeID: strings.Repeat("0", 20), Geo: ptr(0)}, {NodeID: strings.Repeat("0", 20), NodeType: ptr(4)}, {NodeID: strings.Repeat("0", 20), CPU: ptr(16)}, {NodeID: strings.Repeat("0", 20), IPv6: "fe80::1%eth0"}} {
		if _, err := e.Preview(d, false); err == nil {
			t.Fatal("accepted invalid descriptor")
		}
	}
}
func TestDeterminismAndKeyRotation(t *testing.T) {
	e := engine(t)
	d := Descriptor{NodeID: "0123456789abcdef0123", CPU: ptr(4), RAM: ptr(6)}
	a, _ := e.Encode(d, "n", "example.org", false)
	b, _ := e.Encode(d, "p", "example.net", false)
	if a.Label[1:] != b.Label[1:] {
		t.Fatal("cosmetic configuration changed ciphertext")
	}
	key, _ := hex.DecodeString(strings.Repeat("01", 32))
	other, _ := New(key, catalog.Dictionary())
	c, _ := other.Encode(d, "n", "example.org", false)
	if a.Label == c.Label {
		t.Fatal("key change not reflected")
	}
}
func FuzzDecode(f *testing.F) {
	e, _ := New(make([]byte, 32), catalog.Dictionary())
	f.Add("n0abcdefghijklmnopq")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) { _, _ = e.Decode(s) })
}
