package geoipupdate

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCorruptExistingIsNotFallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "city.mmdb")
	os.WriteFile(path, []byte("valid-original"), 0600)
	sum, _ := fileHash(path)
	b, _ := json.Marshal(State{Provider: "DB-IP Lite City", Month: "2026-11", SHA256: sum})
	os.WriteFile(path+".json", b, 0600)
	os.WriteFile(path, []byte("valid-corrupted"), 0600)
	count := 0
	if _, err := testUpdater(nil, 503, &count).run(context.Background(), path); err == nil {
		t.Fatal("corrupt cached file treated as fallback")
	}
}

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func testUpdater(payload []byte, status int, counter *int) *updater {
	return &updater{page: "https://example.test/page", validate: func(path string, _ bool) error {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if !bytes.HasPrefix(b, []byte("valid-")) {
			return errors.New("invalid database")
		}
		return nil
	}, client: &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		body := []byte(`<a href="https://download.db-ip.com/free/dbip-city-lite-2026-11.mmdb.gz">Download</a>`)
		code := 200
		if strings.HasSuffix(r.URL.Path, ".gz") {
			*counter++
			body = payload
			code = status
		}
		return &http.Response{StatusCode: code, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header), ContentLength: int64(len(body))}, nil
	})}}
}
func compressed(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestDownloadAndReuse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "city.mmdb")
	count := 0
	u := testUpdater(compressed(t, "valid-new"), 200, &count)
	r, e := u.run(context.Background(), path)
	if e != nil || r.Status != "updated" || r.SHA256 == "" {
		t.Fatal(r, e)
	}
	r, e = u.run(context.Background(), path)
	if e != nil || r.Status != "current" || count != 1 {
		t.Fatal(r, e, count)
	}
}
func TestFailedUpdateKeepsExisting(t *testing.T) {
	for _, payload := range [][]byte{[]byte("not-gzip"), compressed(t, "invalid-mmdb"), compressed(t, "valid-new")[:10]} {
		path := filepath.Join(t.TempDir(), "city.mmdb")
		if e := os.WriteFile(path, []byte("valid-old"), 0600); e != nil {
			t.Fatal(e)
		}
		count := 0
		r, e := testUpdater(payload, 200, &count).run(context.Background(), path)
		b, _ := os.ReadFile(path)
		if e != nil || r.Status != "kept-existing" || string(b) != "valid-old" {
			t.Fatal(r, e)
		}
	}
}
func TestMissingDatabaseFailure(t *testing.T) {
	count := 0
	path := filepath.Join(t.TempDir(), "city.mmdb")
	if _, e := testUpdater(nil, 503, &count).run(context.Background(), path); e == nil {
		t.Fatal("initial failure hidden")
	}
	if _, e := os.Stat(path); !os.IsNotExist(e) {
		t.Fatal("invalid file installed")
	}
}
func TestUpdateLock(t *testing.T) {
	count := 0
	path := filepath.Join(t.TempDir(), "city.mmdb")
	os.WriteFile(path+".lock", nil, 0600)
	if _, e := testUpdater(nil, 200, &count).run(context.Background(), path); e == nil {
		t.Fatal("parallel updater accepted")
	}
}
