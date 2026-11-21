package codec

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/catalog"
	"net/netip"
	"os"
	"reflect"
	"strings"
	"testing"
)

type frozenVector struct {
	Format     string     `json:"format"`
	KeyHex     string     `json:"keyHex"`
	Descriptor Descriptor `json:"descriptor"`
	Compact    Output     `json:"compact"`
	Padded     Output     `json:"padded"`
}

func frozenVectors(t testing.TB, version string) []frozenVector {
	t.Helper()
	data, err := os.ReadFile("testdata/format" + version + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []frozenVector
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}
func TestFrozenFormatVectors(t *testing.T) {
	for _, version := range []string{"0", "1"} {
		for _, row := range frozenVectors(t, version) {
			key, _ := hex.DecodeString(row.KeyHex)
			e, err := New(key, catalog.Dictionary())
			if err != nil {
				t.Fatal(err)
			}
			for i, want := range []Output{row.Compact, row.Padded} {
				if want.Label[1:2] != version {
					t.Fatal("fixture format mismatch")
				}
				for _, input := range []string{want.Label, want.Hostname, "https://" + strings.ToUpper(want.Hostname) + "/node"} {
					got, err := e.Decode(input)
					if err != nil || !Same(got, row.Descriptor) {
						t.Fatalf("format %s decode: %v", version, err)
					}
				}
				if version == Version {
					got, err := e.Encode(row.Descriptor, "x", "example.org", i == 1)
					if err != nil || !reflect.DeepEqual(got, want) {
						t.Fatalf("format 1 golden encoding changed: %v", err)
					}
				}
			}
		}
	}
}
func TestFrozenFormatParameters(t *testing.T) {
	if Version != "1" || string(tweak) != "PND:format-1:ff1:aes256" || string(legacyTweak) != "PND:format-0:ff1:aes256" {
		t.Fatal("published envelope changed")
	}
	if !reflect.DeepEqual(patterns, []uint64{0, 3, 124, 112, 114, 126, 127, 384, 387, 511, 507, 255, 383, 1, 2, 128}) || !reflect.DeepEqual(widths, []int{32, 16, 2, 3, 4, 4, 3, 32, 128}) {
		t.Fatal("published packing changed")
	}
}
func TestVersionSeparation(t *testing.T) {
	old, current := frozenVectors(t, "0"), frozenVectors(t, "1")
	for i, row := range old {
		if row.Compact.Label[2:] == current[i].Compact.Label[2:] || row.Padded.Label[2:] == current[i].Padded.Label[2:] {
			t.Fatal("format tweaks share ciphertext")
		}
	}
	e := engine(t)
	for _, version := range []string{"2", "9", "a", "z"} {
		if _, err := e.Decode("x" + version + current[0].Compact.Label[2:]); err == nil {
			t.Fatal("unknown version accepted")
		}
	}
}
func TestNoncanonicalPayloads(t *testing.T) {
	e := engine(t)
	d := Descriptor{NodeID: strings.Repeat("0", 20)}
	var w writer
	w.put(1, 1)
	w.put(0, 1)
	w.put(0, 9)
	w.bytes(make([]byte, 10))
	if _, err := e.unpack(w.n.Text(36)); err == nil {
		t.Fatal("raw mask accepted for a known pattern")
	}
	s, _, _, _ := e.pack(d)
	for _, version := range []string{"0", "1"} {
		tw := tweak
		if version == "0" {
			tw = legacyTweak
		}
		ct, err := e.cipher.Encrypt("0"+s, tw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = e.Decode("x" + version + ct); err == nil {
			t.Fatal("noncanonical partial padding accepted")
		}
	}
}
func FuzzRoundTrip(f *testing.F) {
	e, err := New(make([]byte, 32), catalog.Dictionary())
	if err != nil {
		f.Fatal(err)
	}
	f.Add(make([]byte, 64))
	f.Add([]byte(strings.Repeat("\xff", 64)))
	f.Fuzz(func(t *testing.T, data []byte) {
		var b [64]byte
		copy(b[:], data)
		d := Descriptor{NodeID: hex.EncodeToString(b[2:12])}
		mask := binary.BigEndian.Uint16(b[:2]) & 511
		values := []uint32{binary.BigEndian.Uint32(b[12:16]), uint32(binary.BigEndian.Uint16(b[16:18])), uint32(b[18] & 3), uint32(b[19] & 7), uint32(b[20] & 15), uint32(b[21] & 15), uint32(b[22] & 7)}
		if values[0] == 0 {
			values[0] = 1
		}
		if values[3] >= 4 && values[3] <= 6 {
			values[3] = 7
		}
		targets := []**uint32{&d.Geo, &d.Provider, &d.Architecture, &d.NodeType, &d.CPU, &d.RAM, &d.Storage}
		for i, target := range targets {
			if mask&(1<<i) != 0 {
				v := values[i]
				*target = &v
			}
		}
		if mask&128 != 0 {
			d.IPv4 = netip.AddrFrom4([4]byte(b[23:27])).String()
		}
		if mask&256 != 0 {
			ip := [16]byte(b[27:43])
			ip[0] = 0x20
			d.IPv6 = netip.AddrFrom16(ip).String()
		}
		payload, _, _, err := e.pack(d)
		if err != nil {
			t.Fatal(err)
		}
		for _, padded := range []bool{false, true} {
			out, err := e.Encode(d, "x", "example.org", padded)
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Label) > 63 || padded && len(out.Label) != 63 {
				t.Fatal("label length")
			}
			got, err := e.Decode(out.Label)
			if err != nil || !Same(got, d) {
				t.Fatal("format 1 round trip", err)
			}
			plain := payload
			if padded {
				plain = strings.Repeat("0", 61-len(plain)) + plain
			}
			ct, err := e.cipher.Encrypt(plain, legacyTweak)
			if err != nil {
				t.Fatal(err)
			}
			got, err = e.Decode("x0" + ct)
			if err != nil || !Same(got, d) {
				t.Fatal("format 0 round trip", err)
			}
		}
	})
}
