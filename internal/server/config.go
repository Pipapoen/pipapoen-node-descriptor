package server

import (
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/codec"
)

type Config struct {
	SessionTTL                                              time.Duration
	CountryCityEnabled, PlaceSearchEnabled, IPLookupEnabled bool
	BrandName                                               string
	Address, Domain, Prefix, Mode, PublicURL                string
	Key                                                     []byte
	Issuer, ClientID, ClientSecret, GeoNames, GeoIPPath     string
	AllowHTTPOIDC                                           bool
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func featureFlag(name string) (bool, error) {
	value, exists := os.LookupEnv(name)
	if !exists {
		return true, nil
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, errors.New(name + " must be true or false")
	}
	return enabled, nil
}
func sessionDuration(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration < time.Minute || duration > 24*time.Hour || duration%time.Second != 0 {
		return 0, errors.New("PND_SESSION_TTL must be a whole-second duration between 1m and 24h")
	}
	return duration, nil
}
func LoadConfig() (Config, error) {
	c := Config{Address: env("PND_LISTEN", ":8080"), Domain: strings.ToLower(env("PND_DOMAIN", "example.org")), Prefix: env("PND_PREFIX", "x"), Mode: env("PND_LENGTH_MODE", "compact"), PublicURL: strings.TrimSuffix(env("PND_PUBLIC_URL", "http://localhost:8080"), "/"), Issuer: os.Getenv("PND_OIDC_ISSUER"), ClientID: os.Getenv("PND_OIDC_CLIENT_ID"), GeoNames: os.Getenv("PND_GEONAMES_USERNAME"), GeoIPPath: os.Getenv("PND_GEOIP_DB"), AllowHTTPOIDC: os.Getenv("PND_ALLOW_HTTP_OIDC") == "true"}
	c.BrandName = "Pipapoen"
	value, present := os.LookupEnv("PND_SESSION_TTL")
	if !present {
		value = "30m"
	}
	ttl, err := sessionDuration(value)
	if err != nil {
		return c, err
	}
	c.SessionTTL = ttl
	for name, target := range map[string]*bool{"PND_COUNTRY_CITY_ENABLED": &c.CountryCityEnabled, "PND_PLACE_SEARCH_ENABLED": &c.PlaceSearchEnabled, "PND_IP_LOOKUP_ENABLED": &c.IPLookupEnabled} {
		value, err := featureFlag(name)
		if err != nil {
			return c, err
		}
		*target = value
	}
	if brand, ok := os.LookupEnv("PND_BRAND_NAME"); ok {
		c.BrandName = strings.TrimSpace(brand)
	}
	b, err := os.ReadFile(env("PND_KEY_FILE", "secrets/pnd-key"))
	if err != nil {
		return c, errors.New("cannot read PND_KEY_FILE")
	}
	c.Key, err = hex.DecodeString(strings.TrimSpace(string(b)))
	if err != nil || len(c.Key) != 32 {
		return c, errors.New("PND_KEY_FILE must contain 64 hexadecimal characters")
	}
	if p := os.Getenv("PND_OIDC_CLIENT_SECRET_FILE"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return c, errors.New("cannot read OIDC client secret")
		}
		c.ClientSecret = strings.TrimSpace(string(b))
		if c.ClientSecret == "" {
			return c, errors.New("OIDC client secret file is empty")
		}
	}
	return c, c.Validate()
}
func (c Config) Validate() error {
	if c.SessionTTL < time.Minute || c.SessionTTL > 24*time.Hour || c.SessionTTL%time.Second != 0 {
		return errors.New("invalid session duration")
	}
	if len([]rune(c.BrandName)) > 64 || strings.ContainsAny(c.BrandName, "\r\n\x00") {
		return errors.New("PND_BRAND_NAME must be at most 64 characters on one line")
	}
	if len(c.Key) != 32 {
		return errors.New("invalid key")
	}
	if err := codec.ValidateDomain(c.Domain); err != nil {
		return err
	}
	if len(c.Prefix) != 1 || c.Prefix[0] < 'a' || c.Prefix[0] > 'z' {
		return errors.New("PND_PREFIX must be one lowercase letter")
	}
	if c.Mode != "compact" && c.Mode != "padded" {
		return errors.New("invalid length mode")
	}
	u, err := url.Parse(c.PublicURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("PND_PUBLIC_URL must be an HTTP(S) origin")
	}
	if c.Issuer != "" || c.ClientID != "" || c.ClientSecret != "" {
		if c.Issuer == "" || c.ClientID == "" || c.ClientSecret == "" {
			return errors.New("OIDC issuer, client ID and secret are all required")
		}
		i, err := url.Parse(c.Issuer)
		if err != nil || i.Host == "" || i.User != nil || (i.Scheme != "https" && !(c.AllowHTTPOIDC && i.Scheme == "http")) {
			return errors.New("OIDC requires HTTPS or explicit PND_ALLOW_HTTP_OIDC")
		}
	}
	return nil
}
