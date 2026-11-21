package democlient_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/democlient"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/server"
)

func TestBrowserAdapterMatchesDeployedAPI(t *testing.T) {
	client, err := democlient.New()
	if err != nil {
		t.Fatal(err)
	}
	app, err := server.New(context.Background(), server.Config{
		Key: make([]byte, 32), Prefix: "x", Domain: "example.org", Mode: "compact", BrandName: "Pipapoen",
		PublicURL: "http://localhost:8080", SessionTTL: 30 * time.Minute, CountryCityEnabled: true,
	}, fstest.MapFS{})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	requests := []struct{ path, body string }{
		{"/api/session", ""}, {"/api/meta", ""}, {"/api/countries", ""},
		{"/api/cities?country=US&q=New+York", ""}, {"/api/cities?country=TW&q=Taipei", ""},
		{"/api/cities?country=US&offset=100&details=true", ""}, {"/api/cities?country=XX", ""},
		{"/api/cities?country=US&details=invalid", ""}, {"/api/location?id=5128581", ""},
		{"/api/location?id=4294967295", ""}, {"/api/location?id=0", ""},
		{"/api/preview", `{"descriptor":{"nodeId":"0123456789abcdef0123","provider":0,"cpu":0}}`},
		{"/api/encode", `{"descriptor":{"nodeId":"0123456789abcdef0123","geo":5128581,"ipv6":"2001:db8::1234"},"prefix":"z","domain":"nodes.example.org","mode":"padded"}`},
		{"/api/encode", `{"descriptor":{"nodeId":"0123456789abcdef0123","provider":42,"ram":5}}`},
		{"/api/encode", `{"descriptor":{"nodeId":"bad"}}`},
		{"/api/encode", `{"descriptor":{"nodeId":"0123456789abcdef0123","unknown":1}}`},
		{"/api/preview", `{"descriptor":{"nodeId":"0123456789abcdef0123"},"mode":"other"}`},
		{"/api/decode", `{"input":"invalid"}`},
		{"/api/places?q=New+York", ""}, {"/api/geoip?ip=8.8.8.8", ""},
	}
	encoded := client.Request("/api/encode", `{"descriptor":{"nodeId":"0123456789abcdef0123","provider":0,"geo":5128581}}`)
	b, _ := json.Marshal(encoded.Result)
	var output struct {
		Hostname string `json:"hostname"`
	}
	_ = json.Unmarshal(b, &output)
	requests = append(requests, struct{ path, body string }{"/api/decode", `{"input":"` + output.Hostname + `"}`})
	for _, test := range requests {
		t.Run(test.path+test.body, func(t *testing.T) {
			method := "GET"
			if test.body != "" {
				method = "POST"
			}
			response := httptest.NewRecorder()
			app.ServeHTTP(response, httptest.NewRequest(method, test.path, strings.NewReader(test.body)))
			var native any
			if err := json.Unmarshal(response.Body.Bytes(), &native); err != nil {
				t.Fatal(err)
			}
			browser := client.Request(test.path, test.body)
			b, _ := json.Marshal(browser.Result)
			var decoded any
			_ = json.Unmarshal(b, &decoded)
			if test.path == "/api/meta" {
				delete(decoded.(map[string]any), "demo")
			}
			if browser.OK != (response.Code == 200) || !reflect.DeepEqual(native, decoded) {
				t.Fatalf("API parity mismatch: native=%s, browser=%s", response.Body.String(), b)
			}
		})
	}
}
