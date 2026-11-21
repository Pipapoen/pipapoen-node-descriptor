// Package democlient adapts the production codec and reference data for the browser demo.
package democlient

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/buildinfo"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/catalog"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/codec"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/geography"
)

type Result struct {
	OK     bool `json:"ok"`
	Result any  `json:"result"`
}

type Client struct{ engine *codec.Engine }

func New() (*Client, error) {
	// Public demonstration key: 32 zero bytes. Deployment keys are never read here.
	engine, err := codec.New(make([]byte, 32), catalog.Dictionary())
	return &Client{engine}, err
}

func failure(code string) Result { return Result{false, map[string]string{"error": code}} }
func success(value any) Result   { return Result{true, value} }

func parse(body string, target any) bool {
	if len(body) > 16*1024 {
		return false
	}
	d := json.NewDecoder(strings.NewReader(body))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return false
	}
	var extra any
	return d.Decode(&extra) == io.EOF
}

func (c *Client) Request(path, body string) Result {
	u, err := url.ParseRequestURI(path)
	if err != nil {
		return failure("invalid_request")
	}
	query := u.Query()
	switch u.Path {
	case "/api/session":
		return success(map[string]any{"required": false, "authenticated": true, "brandName": "Pipapoen"})
	case "/api/meta":
		return success(map[string]any{"version": buildinfo.Version, "format": codec.Version, "demo": true,
			"brandName": "Pipapoen", "prefix": "x", "domain": "example.org", "mode": "compact",
			"providers": catalog.Providers, "cpu": catalog.CPU, "ram": catalog.RAM, "storage": catalog.Storage,
			"plans": catalog.Plans(), "placesEnabled": false, "geoipEnabled": false, "countriesEnabled": true, "geoipAttribution": ""})
	case "/api/node-id":
		id, err := codec.NewID()
		if err != nil {
			return failure("random_source")
		}
		return success(map[string]string{"nodeId": id})
	case "/api/preview", "/api/encode":
		var request struct {
			Descriptor codec.Descriptor `json:"descriptor"`
			Prefix     string           `json:"prefix"`
			Domain     string           `json:"domain"`
			Mode       string           `json:"mode"`
		}
		if !parse(body, &request) {
			return failure("invalid_request")
		}
		if request.Prefix == "" {
			request.Prefix = "x"
		}
		if request.Domain == "" {
			request.Domain = "example.org"
		}
		if request.Mode == "" {
			request.Mode = "compact"
		}
		if request.Mode != "compact" && request.Mode != "padded" {
			return failure("invalid_request")
		}
		if u.Path == "/api/preview" {
			out, err := c.engine.Preview(request.Descriptor, request.Mode == "padded")
			if err != nil {
				return failure("invalid_descriptor")
			}
			return success(out)
		}
		out, err := c.engine.Encode(request.Descriptor, request.Prefix, request.Domain, request.Mode == "padded")
		if err != nil {
			return failure("invalid_descriptor")
		}
		return success(out)
	case "/api/decode":
		var request struct {
			Input string `json:"input"`
		}
		if !parse(body, &request) {
			return failure("invalid_request")
		}
		out, err := c.engine.Decode(request.Input)
		if err != nil {
			return failure("decode_failed")
		}
		return success(out)
	case "/api/countries":
		return success(geography.Countries())
	case "/api/cities":
		offset := 0
		if value := query.Get("offset"); value != "" {
			offset, err = strconv.Atoi(value)
			if err != nil {
				return failure("invalid_request")
			}
		}
		details := query.Get("details")
		if details != "" && details != "true" && details != "false" {
			return failure("invalid_request")
		}
		out, err := geography.Cities(query.Get("country"), query.Get("q"), offset, details == "true")
		if err != nil {
			return failure("invalid_request")
		}
		return success(out)
	case "/api/location":
		id, err := strconv.ParseUint(query.Get("id"), 10, 32)
		if err != nil || id == 0 {
			return failure("invalid_request")
		}
		out, err := geography.Lookup(uint32(id))
		if err != nil {
			return failure("lookup_failed")
		}
		return success(out)
	default:
		return failure("feature_disabled")
	}
}
