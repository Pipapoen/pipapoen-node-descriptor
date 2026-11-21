package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func testConfig() Config {
	return Config{Address: ":8080", Domain: "example.org", Prefix: "n", Mode: "compact", PublicURL: "http://localhost:8080", Key: make([]byte, 32), CountryCityEnabled: true, PlaceSearchEnabled: true, IPLookupEnabled: true, SessionTTL: 30 * time.Minute}
}

func TestFeatureFlags(t *testing.T) {
	for _, mask := range []int{0, 1, 2, 3, 4, 5, 6, 7} {
		cfg := testConfig()
		cfg.CountryCityEnabled = mask&1 != 0
		cfg.PlaceSearchEnabled = mask&2 != 0
		cfg.IPLookupEnabled = mask&4 != 0
		if !cfg.IPLookupEnabled {
			cfg.GeoIPPath = "missing-disabled-database.mmdb"
		}
		app, err := New(context.Background(), cfg, fstest.MapFS{"index.html": {Data: []byte("PND")}})
		if err != nil {
			t.Fatal(err)
		}
		tests := []struct {
			path          string
			enabled       bool
			enabledStatus int
		}{{"/api/countries", cfg.CountryCityEnabled, 200}, {"/api/cities?country=US", cfg.CountryCityEnabled, 200}, {"/api/places?q=New%20York", cfg.PlaceSearchEnabled, 503}, {"/api/geoip?ip=8.8.8.8", cfg.IPLookupEnabled, 503}}
		for _, test := range tests {
			want := 404
			if test.enabled {
				want = test.enabledStatus
			}
			if got := requestApp(app, "GET", test.path, "").Code; got != want {
				t.Fatalf("mask %d %s: got %d want %d", mask, test.path, got, want)
			}
		}
		if requestApp(app, "GET", "/api/location?id=5128581", "").Code != 200 {
			t.Fatal("decode labels disabled by helper flag")
		}
		if requestApp(app, "POST", "/api/encode", `{"descriptor":{"nodeId":"0123456789abcdef0123","geo":5128581}}`).Code != 200 {
			t.Fatal("manual geography disabled")
		}
	}
}
func TestFeatureFlagParsing(t *testing.T) {
	for _, value := range []string{"true", "false", "invalid", ""} {
		t.Setenv("PND_TEST_FLAG", value)
		v, err := featureFlag("PND_TEST_FLAG")
		if value == "true" || value == "false" {
			if err != nil || v != (value == "true") {
				t.Fatal("valid flag rejected")
			}
		} else if err == nil {
			t.Fatal("invalid flag accepted")
		}
	}
}

func TestBrandAndGeographyAPI(t *testing.T) {
	cfg := testConfig()
	cfg.BrandName = "Example Lab"
	app, err := New(context.Background(), cfg, fstest.MapFS{"index.html": {Data: []byte("PND")}})
	if err != nil {
		t.Fatal(err)
	}
	meta := requestApp(app, "GET", "/api/meta", "")
	if !bytes.Contains(meta.Body.Bytes(), []byte(`"brandName":"Example Lab"`)) {
		t.Fatal("brand missing")
	}
	if requestApp(app, "GET", "/api/countries", "").Code != 200 {
		t.Fatal("countries unavailable")
	}
	location := requestApp(app, "GET", "/api/location?id=5128581", "")
	if location.Code != 200 || !bytes.Contains(location.Body.Bytes(), []byte(`"code":"US"`)) {
		t.Fatal("ID lookup failed")
	}
	cities := requestApp(app, "GET", "/api/cities?country=TW&q=Taipei", "")
	if cities.Code != 200 || !bytes.Contains(cities.Body.Bytes(), []byte(`"id":1668341`)) {
		t.Fatal("city lookup failed", cities.Body.String())
	}
	if requestApp(app, "GET", "/api/cities?country=US&offset=-1", "").Code != 400 {
		t.Fatal("invalid offset accepted")
	}
	cfg.BrandName = strings.Repeat("a", 65)
	if cfg.Validate() == nil {
		t.Fatal("oversized brand accepted")
	}
}
func testApp(t *testing.T) *App {
	t.Helper()
	a, err := New(context.Background(), testConfig(), fstest.MapFS{"index.html": {Data: []byte("PND")}})
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func requestApp(a *App, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)
	return w
}
func TestPublicAPI(t *testing.T) {
	a := testApp(t)
	w := requestApp(a, "POST", "/api/encode", `{"descriptor":{"nodeId":"0123456789abcdef0123","cpu":4},"mode":"padded"}`)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var out struct {
		Hostname string `json:"hostname"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	payload, _ := json.Marshal(map[string]string{"input": out.Hostname})
	w = requestApp(a, "POST", "/api/decode", string(payload))
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"cpu":4`)) {
		t.Fatal(w.Code, w.Body.String())
	}
	if requestApp(a, "GET", "/api/places?q=Taipei", "").Code != 503 {
		t.Fatal("unconfigured lookup")
	}
}
func TestRejectUnknownFieldsAndCrossOrigin(t *testing.T) {
	a := testApp(t)
	if requestApp(a, "POST", "/api/preview", `{"extra":true}`).Code != 400 {
		t.Fatal("unknown fields accepted")
	}
	r := httptest.NewRequest("POST", "/api/node-id", strings.NewReader("{}"))
	r.Header.Set("Origin", "https://attacker.invalid")
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross origin accepted")
	}
}
func TestOIDCFailsClosed(t *testing.T) {
	c := testConfig()
	c.ClientID = "test"
	if c.Validate() == nil {
		t.Fatal("partial OIDC accepted")
	}
	c = testConfig()
	c.Issuer = "http://issuer.invalid"
	c.ClientID = "test"
	c.ClientSecret = "test"
	if c.Validate() == nil {
		t.Fatal("implicit insecure issuer accepted")
	}
}
func TestSessionCookie(t *testing.T) {
	a := &auth{signing: []byte("test-only-signing-key"), sessionTTL: 30 * time.Minute}
	signed := a.sign(ticket{IssuedAt: time.Now().Unix(), Expiry: time.Now().Add(time.Minute).Unix(), Session: true})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "pnd_session", Value: signed})
	if !a.logged(r) {
		t.Fatal("valid cookie rejected")
	}
	r = httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "pnd_session", Value: signed + "x"})
	if a.logged(r) {
		t.Fatal("tampered cookie accepted")
	}
	r = httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "pnd_session", Value: a.sign(ticket{Expiry: time.Now().Add(-time.Second).Unix(), Session: true})})
	if a.logged(r) {
		t.Fatal("expired cookie accepted")
	}
}
