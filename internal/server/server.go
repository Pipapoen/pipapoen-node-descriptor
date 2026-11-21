package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/buildinfo"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/catalog"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/codec"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/geography"
	"github.com/oschwald/geoip2-golang/v2"
	"golang.org/x/time/rate"
)

type App struct {
	Config  Config
	Codec   *codec.Engine
	auth    *auth
	geo     *geoip2.Reader
	handler http.Handler
}

func New(ctx context.Context, c Config, assets fs.FS) (*App, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	engine, err := codec.New(c.Key, catalog.Dictionary())
	if err != nil {
		return nil, err
	}
	a, err := newAuth(ctx, c)
	if err != nil {
		return nil, err
	}
	app := &App{Config: c, Codec: engine, auth: a}
	if c.IPLookupEnabled && c.GeoIPPath != "" {
		app.geo, err = geoip2.Open(c.GeoIPPath)
		if err != nil {
			return nil, errors.New("cannot open GeoIP database")
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "version": buildinfo.Version})
	})
	mux.HandleFunc("GET /api/session", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"required": a != nil, "authenticated": a.logged(r), "brandName": c.BrandName})
	})
	if a != nil {
		mux.HandleFunc("GET /auth/login", a.login)
		mux.HandleFunc("GET /auth/callback", a.callback)
		mux.HandleFunc("POST /auth/logout", func(w http.ResponseWriter, r *http.Request) {
			if !sameOrigin(r) {
				http.Error(w, "Origin rejected", 403)
				return
			}
			a.cookie(w, "pnd_session", "", -1)
			writeJSON(w, map[string]bool{"ok": true})
		})
	}
	api := http.NewServeMux()
	api.HandleFunc("GET /api/meta", app.meta)
	api.HandleFunc("POST /api/node-id", func(w http.ResponseWriter, r *http.Request) {
		id, err := codec.NewID()
		if err != nil {
			fail(w, "random_source", 500)
			return
		}
		writeJSON(w, map[string]string{"nodeId": id})
	})
	api.HandleFunc("POST /api/preview", app.preview)
	api.HandleFunc("POST /api/encode", app.encode)
	api.HandleFunc("POST /api/decode", app.decode)
	api.HandleFunc("GET /api/places", app.places)
	api.HandleFunc("GET /api/geoip", app.geoip)
	api.HandleFunc("GET /api/countries", func(w http.ResponseWriter, r *http.Request) {
		if !c.CountryCityEnabled {
			fail(w, "feature_disabled", 404)
			return
		}
		writeJSON(w, geography.Countries())
	})
	api.HandleFunc("GET /api/location", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
		if err != nil || id == 0 {
			fail(w, "invalid_request", 400)
			return
		}
		result, err := geography.Lookup(uint32(id))
		if err != nil {
			fail(w, "lookup_failed", 500)
			return
		}
		writeJSON(w, result)
	})
	api.HandleFunc("GET /api/cities", func(w http.ResponseWriter, r *http.Request) {
		if !c.CountryCityEnabled {
			fail(w, "feature_disabled", 404)
			return
		}
		offset := 0
		var err error
		if raw := r.URL.Query().Get("offset"); raw != "" {
			offset, err = strconv.Atoi(raw)
			if err != nil {
				fail(w, "invalid_request", 400)
				return
			}
		}
		details := r.URL.Query().Get("details")
		if details != "" && details != "true" && details != "false" {
			fail(w, "invalid_request", 400)
			return
		}
		out, err := geography.Cities(r.URL.Query().Get("country"), r.URL.Query().Get("q"), offset, details == "true")
		if err != nil {
			fail(w, "invalid_request", 400)
			return
		}
		writeJSON(w, out)
	})
	limiter := rate.NewLimiter(30, 60)
	mux.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.logged(r) {
			fail(w, "login_required", 401)
			return
		}
		if r.Method == "POST" && !sameOrigin(r) {
			fail(w, "origin_rejected", 403)
			return
		}
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			fail(w, "rate_limit", 429)
			return
		}
		api.ServeHTTP(w, r)
	}))
	mux.Handle("/", http.FileServerFS(assets))
	app.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		mux.ServeHTTP(w, r)
	})
	return app, nil
}
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) { a.handler.ServeHTTP(w, r) }
func (a *App) Close() {
	if a.geo != nil {
		_ = a.geo.Close()
	}
}
func sameOrigin(r *http.Request) bool {
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, e := url.Parse(origin)
	return e == nil && u.Host == r.Host && (u.Scheme == "http" || u.Scheme == "https")
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
func body(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		fail(w, "invalid_request", 400)
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		fail(w, "invalid_request", 400)
		return false
	}
	return true
}

type request struct {
	Descriptor codec.Descriptor `json:"descriptor"`
	Prefix     string           `json:"prefix"`
	Domain     string           `json:"domain"`
	Mode       string           `json:"mode"`
}

func (a *App) defaults(v *request) error {
	if v.Prefix == "" {
		v.Prefix = a.Config.Prefix
	}
	if v.Domain == "" {
		v.Domain = a.Config.Domain
	}
	if v.Mode == "" {
		v.Mode = a.Config.Mode
	}
	if v.Mode != "compact" && v.Mode != "padded" {
		return errors.New("mode")
	}
	return nil
}
func (a *App) meta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"version": buildinfo.Version, "format": codec.Version, "brandName": a.Config.BrandName, "prefix": a.Config.Prefix, "domain": a.Config.Domain, "mode": a.Config.Mode, "providers": catalog.Providers, "cpu": catalog.CPU, "ram": catalog.RAM, "storage": catalog.Storage, "plans": catalog.Plans(), "catalog": catalog.Info(), "placesEnabled": a.Config.PlaceSearchEnabled && a.Config.GeoNames != "", "geoipEnabled": a.geo != nil, "countriesEnabled": a.Config.CountryCityEnabled, "geoipAttribution": a.geoipAttribution()})
}
func (a *App) geoipAttribution() string {
	if a.geo == nil {
		return ""
	}
	if strings.HasPrefix(a.geo.Metadata().DatabaseType, "DBIP") {
		return "DB-IP"
	}
	return "MaxMind"
}
func (a *App) preview(w http.ResponseWriter, r *http.Request) {
	var v request
	if !body(w, r, &v) {
		return
	}
	if a.defaults(&v) != nil {
		fail(w, "invalid_request", 400)
		return
	}
	out, err := a.Codec.Preview(v.Descriptor, v.Mode == "padded")
	if err != nil {
		fail(w, "invalid_descriptor", 400)
		return
	}
	writeJSON(w, out)
}
func (a *App) encode(w http.ResponseWriter, r *http.Request) {
	var v request
	if !body(w, r, &v) {
		return
	}
	if a.defaults(&v) != nil {
		fail(w, "invalid_request", 400)
		return
	}
	out, err := a.Codec.Encode(v.Descriptor, v.Prefix, v.Domain, v.Mode == "padded")
	if err != nil {
		fail(w, "invalid_descriptor", 400)
		return
	}
	writeJSON(w, out)
}
func (a *App) decode(w http.ResponseWriter, r *http.Request) {
	var v struct {
		Input string `json:"input"`
	}
	if !body(w, r, &v) {
		return
	}
	out, err := a.Codec.Decode(v.Input)
	if err != nil {
		fail(w, "decode_failed", 400)
		return
	}
	writeJSON(w, out)
}
func (a *App) places(w http.ResponseWriter, r *http.Request) {
	if !a.Config.PlaceSearchEnabled {
		fail(w, "feature_disabled", 404)
		return
	}
	if a.Config.GeoNames == "" {
		fail(w, "places_unavailable", 503)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	id := r.URL.Query().Get("id")
	if len(q) > 120 || len(q) < 2 && id == "" {
		fail(w, "invalid_request", 400)
		return
	}
	endpoint := "https://secure.geonames.org/searchJSON"
	args := url.Values{"username": {a.Config.GeoNames}, "maxRows": {"8"}, "featureClass": {"P"}, "q": {q}}
	if id != "" {
		if _, err := strconv.ParseUint(id, 10, 32); err != nil {
			fail(w, "invalid_request", 400)
			return
		}
		endpoint = "https://secure.geonames.org/getJSON"
		args = url.Values{"username": {a.Config.GeoNames}, "geonameId": {id}}
	}
	req, err := http.NewRequestWithContext(r.Context(), "GET", endpoint+"?"+args.Encode(), nil)
	if err != nil {
		fail(w, "lookup_failed", 502)
		return
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		fail(w, "lookup_failed", 502)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fail(w, "lookup_failed", 502)
		return
	}
	var data struct {
		Geonames []json.RawMessage `json:"geonames"`
		Status   json.RawMessage   `json:"status"`
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil || json.Unmarshal(b, &data) != nil || len(data.Status) > 0 {
		fail(w, "lookup_failed", 502)
		return
	}
	if id != "" {
		writeJSON(w, json.RawMessage(b))
	} else {
		writeJSON(w, data.Geonames)
	}
}
func (a *App) geoip(w http.ResponseWriter, r *http.Request) {
	if !a.Config.IPLookupEnabled {
		fail(w, "feature_disabled", 404)
		return
	}
	if a.geo == nil {
		fail(w, "geoip_unavailable", 503)
		return
	}
	ip, err := netip.ParseAddr(r.URL.Query().Get("ip"))
	if err != nil || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.Zone() != "" {
		fail(w, "public_ip_required", 400)
		return
	}
	record, err := a.geo.City(ip)
	if err != nil {
		fail(w, "lookup_failed", 502)
		return
	}
	admin := ""
	if len(record.Subdivisions) > 0 {
		admin = record.Subdivisions[0].Names.English
	}
	candidates := []geography.City{}
	if record.City.GeoNameID == 0 {
		candidates, _ = geography.MatchCity(record.Country.ISOCode, record.City.Names.English, admin)
	}
	writeJSON(w, map[string]any{"ip": ip.String(), "geo": record.City.GeoNameID, "name": record.City.Names.English, "country": record.Country.ISOCode, "admin": admin, "accuracyRadiusKm": record.Location.AccuracyRadius, "candidates": candidates, "source": a.geoipAttribution()})
}
