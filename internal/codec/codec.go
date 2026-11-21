package codec

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"net/url"
	"sort"
	"strings"

	"github.com/joshuajorel/scrambler/ff1"
)

// Version is the default encoding format. Published formats remain decodable.
const Version = "1"

var alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
var patterns = []uint64{0, 3, 124, 112, 114, 126, 127, 384, 387, 511, 507, 255, 383, 1, 2, 128}
var widths = []int{32, 16, 2, 3, 4, 4, 3, 32, 128}
var tweak = []byte("PND:format-1:ff1:aes256")
var legacyTweak = []byte("PND:format-0:ff1:aes256")

type Descriptor struct {
	NodeID       string  `json:"nodeId"`
	Geo          *uint32 `json:"geo,omitempty"`
	Provider     *uint32 `json:"provider,omitempty"`
	Architecture *uint32 `json:"architecture,omitempty"`
	NodeType     *uint32 `json:"nodeType,omitempty"`
	CPU          *uint32 `json:"cpu,omitempty"`
	RAM          *uint32 `json:"ram,omitempty"`
	Storage      *uint32 `json:"storage,omitempty"`
	IPv4         string  `json:"ipv4,omitempty"`
	IPv6         string  `json:"ipv6,omitempty"`
}
type Preview struct {
	Bits          int    `json:"bits"`
	CapacityBits  int    `json:"capacityBits"`
	CompactLength int    `json:"compactLength"`
	LabelLength   int    `json:"labelLength"`
	Mode          string `json:"mode"`
}
type Output struct {
	Descriptor Descriptor `json:"descriptor"`
	Label      string     `json:"label"`
	Hostname   string     `json:"hostname"`
	Preview    Preview    `json:"preview"`
}
type Engine struct {
	cipher      *ff1.Cipher
	table       [][3]uint32
	projections [8][][]uint32
}

func New(key []byte, table [][3]uint32) (*Engine, error) {
	if len(key) != 32 {
		return nil, errors.New("a 32-byte key is required")
	}
	c, err := ff1.New(key, ff1.LowerAlphanumeric, ff1.WithMinLength(17), ff1.WithMaxLength(61), ff1.WithMaxTweakLength(64))
	if err != nil {
		return nil, err
	}
	e := &Engine{cipher: c, table: table}
	for mask := 1; mask < 8; mask++ {
		selected := []int{}
		for i := 0; i < 3; i++ {
			if mask&(1<<i) != 0 {
				selected = append(selected, i)
			}
		}
		e.projections[mask] = e.buildProjection(selected)
	}
	return e, nil
}
func NewID() (string, error) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
func (d Descriptor) values() []*uint32 {
	return []*uint32{d.Geo, d.Provider, d.Architecture, d.NodeType, d.CPU, d.RAM, d.Storage}
}
func (d *Descriptor) validate() error {
	id, err := hex.DecodeString(d.NodeID)
	if err != nil || len(id) != 10 {
		return errors.New("nodeId must contain exactly 20 hexadecimal characters")
	}
	d.NodeID = strings.ToLower(d.NodeID)
	for i, p := range d.values() {
		if p != nil && uint64(*p) >= uint64(1)<<widths[i] {
			return errors.New("field value exceeds its allowed range")
		}
	}
	if d.Geo != nil && *d.Geo == 0 {
		return errors.New("GeoNames ID must be positive")
	}
	if d.NodeType != nil && *d.NodeType >= 4 && *d.NodeType <= 6 {
		return errors.New("reserved node type")
	}
	if d.IPv4 != "" {
		a, e := netip.ParseAddr(d.IPv4)
		if e != nil || !a.Is4() || a.Zone() != "" {
			return errors.New("invalid IPv4 address")
		}
		d.IPv4 = a.String()
	}
	if d.IPv6 != "" {
		a, e := netip.ParseAddr(d.IPv6)
		if e != nil || !a.Is6() || a.Is4In6() || a.Zone() != "" {
			return errors.New("invalid IPv6 address")
		}
		d.IPv6 = a.String()
	}
	return nil
}

type writer struct {
	n    big.Int
	size int
}

func (w *writer) put(v uint64, n int) {
	w.n.Lsh(&w.n, uint(n))
	var x big.Int
	x.SetUint64(v)
	w.n.Or(&w.n, &x)
	w.size += n
}
func (w *writer) bytes(b []byte) {
	for _, v := range b {
		w.put(uint64(v), 8)
	}
}
func (w *writer) variable(v uint32) {
	for v >= 128 {
		w.put(uint64(byte(v)|128), 8)
		v >>= 7
	}
	w.put(uint64(v), 8)
}

type reader struct {
	s string
	p int
}

func (r *reader) get(n int) (uint64, error) {
	if r.p+n > len(r.s) {
		return 0, errors.New("truncated payload")
	}
	var v uint64
	for _, c := range r.s[r.p : r.p+n] {
		v = v*2 + uint64(c-'0')
	}
	r.p += n
	return v, nil
}
func (r *reader) bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	for i := range b {
		v, e := r.get(8)
		if e != nil {
			return nil, e
		}
		b[i] = byte(v)
	}
	return b, nil
}
func (r *reader) variable(maxBits int) (uint32, error) {
	var value uint64
	for shift := 0; shift < maxBits+7; shift += 7 {
		b, e := r.get(8)
		if e != nil {
			return 0, e
		}
		value |= (b & 127) << shift
		if b&128 == 0 {
			if value >= uint64(1)<<maxBits {
				return 0, errors.New("integer overflow")
			}
			return uint32(value), nil
		}
	}
	return 0, errors.New("oversized integer")
}
func bitWidth(count int) int {
	n := 0
	for count > 1 {
		n++
		count = (count + 1) / 2
	}
	return n
}
func (e *Engine) projection(selected []int) [][]uint32 {
	mask := 0
	for _, i := range selected {
		mask |= 1 << i
	}
	return e.projections[mask]
}
func (e *Engine) buildProjection(selected []int) [][]uint32 {
	seen := map[string]bool{}
	out := [][]uint32{}
	for _, row := range e.table {
		values := []uint32{}
		for _, i := range selected {
			values = append(values, row[i])
		}
		key := fmt.Sprint(values)
		if !seen[key] {
			seen[key] = true
			out = append(out, values)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		for k := range out[i] {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return false
	})
	return out
}
func (e *Engine) pack(d Descriptor) (string, int, string, error) {
	if err := d.validate(); err != nil {
		return "", 0, "", err
	}
	vals := d.values()
	var mask uint64
	for i, p := range vals {
		if p != nil {
			mask |= 1 << i
		}
	}
	if d.IPv4 != "" {
		mask |= 128
	}
	if d.IPv6 != "" {
		mask |= 256
	}
	var w writer
	w.put(1, 1)
	pi := -1
	for i, m := range patterns {
		if m == mask {
			pi = i
			break
		}
	}
	if pi >= 0 {
		w.put(1, 1)
		w.put(uint64(pi), 4)
	} else {
		w.put(0, 1)
		w.put(mask, 9)
	}
	id, _ := hex.DecodeString(d.NodeID)
	w.bytes(id)
	fixed, varBits := 0, 0
	for i := 0; i < 2; i++ {
		if vals[i] != nil {
			fixed += widths[i]
			v := *vals[i]
			varBits += 8
			for v >= 128 {
				varBits += 8
				v >>= 7
			}
		}
	}
	if fixed > 0 {
		variable := varBits < fixed
		if variable {
			w.put(1, 1)
		} else {
			w.put(0, 1)
		}
		for i := 0; i < 2; i++ {
			if vals[i] != nil {
				if variable {
					w.variable(*vals[i])
				} else {
					w.put(uint64(*vals[i]), widths[i])
				}
			}
		}
	}
	for i := 2; i < 4; i++ {
		if vals[i] != nil {
			w.put(uint64(*vals[i]), widths[i])
		}
	}
	selected := []int{}
	values := []uint32{}
	rawBits := 0
	for i := 4; i < 7; i++ {
		if vals[i] != nil {
			selected = append(selected, i-4)
			values = append(values, *vals[i])
			rawBits += widths[i]
		}
	}
	mode := "raw"
	if len(selected) > 0 {
		table := e.projection(selected)
		index := -1
		for i, row := range table {
			if fmt.Sprint(row) == fmt.Sprint(values) {
				index = i
				break
			}
		}
		if index >= 0 && bitWidth(len(table)) < rawBits {
			mode = "dictionary"
			w.put(1, 1)
			w.put(uint64(index), bitWidth(len(table)))
		} else {
			w.put(0, 1)
			for i := 4; i < 7; i++ {
				if vals[i] != nil {
					w.put(uint64(*vals[i]), widths[i])
				}
			}
		}
	}
	if d.IPv4 != "" {
		a := netip.MustParseAddr(d.IPv4).As4()
		w.bytes(a[:])
	}
	if d.IPv6 != "" {
		a := netip.MustParseAddr(d.IPv6).As16()
		w.bytes(a[:])
	}
	return w.n.Text(36), w.size, mode, nil
}
func (e *Engine) unpack(s string) (Descriptor, error) {
	var d Descriptor
	n, ok := new(big.Int).SetString(s, 36)
	if !ok || n.Sign() <= 0 {
		return d, errors.New("invalid payload")
	}
	r := reader{s: n.Text(2)[1:]}
	flag, err := r.get(1)
	if err != nil {
		return d, err
	}
	var mask uint64
	if flag == 1 {
		v, er := r.get(4)
		if er != nil {
			return d, er
		}
		mask = patterns[v]
	} else {
		mask, err = r.get(9)
		if err != nil {
			return d, err
		}
	}
	id, err := r.bytes(10)
	if err != nil {
		return d, err
	}
	d.NodeID = hex.EncodeToString(id)
	vals := make([]*uint32, 7)
	for i := range vals {
		if mask&(1<<i) != 0 {
			v := uint32(0)
			vals[i] = &v
		}
	}
	if mask&3 != 0 {
		flag, err = r.get(1)
		if err != nil {
			return d, err
		}
		for i := 0; i < 2; i++ {
			if vals[i] != nil {
				if flag == 1 {
					v, er := r.variable(widths[i])
					if er != nil {
						return d, er
					}
					*vals[i] = v
				} else {
					v, er := r.get(widths[i])
					if er != nil {
						return d, er
					}
					*vals[i] = uint32(v)
				}
			}
		}
	}
	for i := 2; i < 4; i++ {
		if vals[i] != nil {
			v, er := r.get(widths[i])
			if er != nil {
				return d, er
			}
			*vals[i] = uint32(v)
		}
	}
	selected := []int{}
	for i := 4; i < 7; i++ {
		if vals[i] != nil {
			selected = append(selected, i-4)
		}
	}
	if len(selected) > 0 {
		flag, err = r.get(1)
		if err != nil {
			return d, err
		}
		if flag == 1 {
			table := e.projection(selected)
			if len(table) == 0 {
				return d, errors.New("missing dictionary")
			}
			v, er := r.get(bitWidth(len(table)))
			if er != nil || int(v) >= len(table) {
				return d, errors.New("invalid dictionary index")
			}
			for j, i := range selected {
				*vals[i+4] = table[v][j]
			}
		} else {
			for i := 4; i < 7; i++ {
				if vals[i] != nil {
					v, er := r.get(widths[i])
					if er != nil {
						return d, er
					}
					*vals[i] = uint32(v)
				}
			}
		}
	}
	d.Geo, d.Provider, d.Architecture, d.NodeType, d.CPU, d.RAM, d.Storage = vals[0], vals[1], vals[2], vals[3], vals[4], vals[5], vals[6]
	if mask&128 != 0 {
		b, er := r.bytes(4)
		if er != nil {
			return d, er
		}
		d.IPv4 = netip.AddrFrom4([4]byte(b)).String()
	}
	if mask&256 != 0 {
		b, er := r.bytes(16)
		if er != nil {
			return d, er
		}
		d.IPv6 = netip.AddrFrom16([16]byte(b)).String()
	}
	if r.p != len(r.s) {
		return d, errors.New("trailing data")
	}
	if err := d.validate(); err != nil {
		return d, err
	}
	canonical, _, _, err := e.pack(d)
	if err != nil || canonical != n.Text(36) {
		return d, errors.New("noncanonical data")
	}
	return d, nil
}
func ValidateDomain(domain string) error {
	if domain == "" || len(domain) > 189 {
		return errors.New("domain must contain 1 to 189 ASCII characters")
	}
	for _, label := range strings.Split(domain, ".") {
		if len(label) < 1 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return errors.New("invalid domain")
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return errors.New("use an ASCII domain")
			}
		}
	}
	return nil
}
func (e *Engine) Preview(d Descriptor, padded bool) (Preview, error) {
	s, b, m, err := e.pack(d)
	if err != nil {
		return Preview{}, err
	}
	l := len(s) + 2
	if l > 63 {
		return Preview{}, errors.New("selection exceeds hostname capacity")
	}
	out := l
	if padded {
		out = 63
	}
	return Preview{b, 315, l, out, m}, nil
}
func (e *Engine) Encode(d Descriptor, prefix, domain string, padded bool) (Output, error) {
	if len(prefix) != 1 || prefix[0] < 'a' || prefix[0] > 'z' {
		return Output{}, errors.New("prefix must be one lowercase letter")
	}
	if err := ValidateDomain(domain); err != nil {
		return Output{}, err
	}
	if err := d.validate(); err != nil {
		return Output{}, err
	}
	p, err := e.Preview(d, padded)
	if err != nil {
		return Output{}, err
	}
	s, _, _, _ := e.pack(d)
	if padded {
		s = strings.Repeat("0", 61-len(s)) + s
	}
	ct, err := e.cipher.Encrypt(s, tweak)
	if err != nil {
		return Output{}, errors.New("encryption failed")
	}
	label := prefix + Version + ct
	return Output{d, label, label + "." + domain, p}, nil
}
func (e *Engine) Decode(input string) (Descriptor, error) {
	input = strings.TrimSpace(input)
	if len(input) > 2048 {
		return Descriptor{}, errors.New("input too long")
	}
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err != nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return Descriptor{}, errors.New("invalid URL")
		}
		input = u.Hostname()
	}
	label := strings.ToLower(strings.Split(strings.TrimSuffix(input, "."), ".")[0])
	if len(label) < 19 || len(label) > 63 || label[0] < 'a' || label[0] > 'z' {
		return Descriptor{}, errors.New("invalid label length or prefix")
	}
	var decodingTweak []byte
	switch label[1:2] {
	case "0":
		decodingTweak = legacyTweak
	case Version:
		decodingTweak = tweak
	default:
		return Descriptor{}, errors.New("unsupported format version")
	}
	for _, c := range label[2:] {
		if !strings.ContainsRune(alphabet, c) {
			return Descriptor{}, errors.New("invalid label character")
		}
	}
	pt, err := e.cipher.Decrypt(label[2:], decodingTweak)
	if err != nil {
		return Descriptor{}, errors.New("cannot decode label")
	}
	short := strings.TrimLeft(pt, "0")
	if short == "" || pt != short && len(pt) != 61 {
		return Descriptor{}, errors.New("invalid padding")
	}
	d, err := e.unpack(short)
	if err != nil {
		return Descriptor{}, errors.New("invalid descriptor or mismatched key")
	}
	return d, nil
}
func Same(a, b Descriptor) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}
