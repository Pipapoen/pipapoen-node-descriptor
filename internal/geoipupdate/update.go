// Package geoipupdate downloads an attributed DB-IP Lite snapshot as an explicit one-shot job.
package geoipupdate

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang/v2"
)

const sourcePage = "https://db-ip.com/db/download/ip-to-city-lite"
const maxCompressed = 128 << 20
const maxDatabase = 512 << 20

var downloadLink = regexp.MustCompile(`https://download\.db-ip\.com/free/dbip-city-lite-(\d{4}-\d{2})\.mmdb\.gz`)

type State struct {
	Provider     string    `json:"provider"`
	Month        string    `json:"month"`
	SHA256       string    `json:"sha256"`
	DownloadedAt time.Time `json:"downloadedAt"`
	License      string    `json:"license"`
}
type Result struct {
	Status  string `json:"status"`
	Month   string `json:"month,omitempty"`
	SHA256  string `json:"sha256,omitempty"`
	Warning string `json:"warning,omitempty"`
}
type updater struct {
	client   *http.Client
	page     string
	validate func(string, bool) error
}

func Update(ctx context.Context, path string) (Result, error) {
	client := &http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || len(via) > 5 {
			return errors.New("unsupported redirect")
		}
		return nil
	}}
	return (&updater{client: client, page: sourcePage, validate: validate}).run(ctx, path)
}
func validate(path string, full bool) error {
	db, err := maxminddb.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()
	if db.Metadata.DatabaseType != "DBIP-City-Lite" || db.Metadata.IPVersion != 6 {
		return errors.New("expected DB-IP Lite City IPv4/IPv6 database")
	}
	if full {
		return db.Verify()
	}
	return nil
}
func (u *updater) run(ctx context.Context, path string) (Result, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Result{}, err
	}
	// Serialize writers. A lock left by a killed updater requires an operator check.
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return Result{}, errors.New("another update is running or a stale update lock exists")
	}
	lock.Close()
	defer os.Remove(path + ".lock")
	existing := u.validate(path, false) == nil
	var state State
	if b, err := os.ReadFile(path + ".json"); err == nil {
		_ = json.Unmarshal(b, &state)
	}
	if existing {
		if state.SHA256 != "" {
			sum, err := fileHash(path)
			existing = err == nil && sum == state.SHA256
		} else {
			existing = u.validate(path, true) == nil
		}
	}
	fallback := func(reason string, err error) (Result, error) {
		if existing {
			return Result{Status: "kept-existing", Warning: reason}, nil
		}
		return Result{}, err
	}
	request, err := http.NewRequestWithContext(ctx, "GET", u.page, nil)
	if err != nil {
		return fallback("source request failed", err)
	}
	response, err := u.client.Do(request)
	if err != nil {
		return fallback("source unavailable", err)
	}
	page, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	response.Body.Close()
	if response.StatusCode != 200 || readErr != nil {
		return fallback("source page unavailable", errors.New("cannot read source page"))
	}
	match := downloadLink.FindSubmatch(page)
	if len(match) != 2 {
		return fallback("download link unavailable", errors.New("cannot locate official MMDB download"))
	}
	source, month := string(match[0]), string(match[1])
	if existing && state.Month == month && state.Provider == "DB-IP Lite City" {
		sum, err := fileHash(path)
		if err == nil && sum == state.SHA256 {
			return Result{Status: "current", Month: month, SHA256: sum}, nil
		}
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".geoip-download-*")
	if err != nil {
		return fallback("cannot create download file", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	request, err = http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		temp.Close()
		return fallback("download request failed", err)
	}
	response, err = u.client.Do(request)
	if err != nil {
		temp.Close()
		return fallback("download unavailable", err)
	}
	if response.StatusCode != 200 {
		response.Body.Close()
		temp.Close()
		return fallback("download unavailable", fmt.Errorf("download returned HTTP %d", response.StatusCode))
	}
	if response.ContentLength > maxCompressed {
		response.Body.Close()
		temp.Close()
		return fallback("download exceeds size limit", errors.New("compressed file too large"))
	}
	compressed := &io.LimitedReader{R: response.Body, N: maxCompressed + 1}
	archive, err := gzip.NewReader(compressed)
	if err != nil {
		response.Body.Close()
		temp.Close()
		return fallback("invalid compressed download", err)
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(temp, hash), io.LimitReader(archive, maxDatabase+1))
	closeErr := archive.Close()
	response.Body.Close()
	syncErr := temp.Sync()
	fileErr := temp.Close()
	if copyErr != nil || closeErr != nil || syncErr != nil || fileErr != nil || written > maxDatabase || compressed.N <= 0 {
		return fallback("download validation failed", errors.New("invalid, incomplete, or oversized database"))
	}
	if err := u.validate(tempName, true); err != nil {
		return fallback("MMDB validation failed", err)
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	state = State{Provider: "DB-IP Lite City", Month: month, SHA256: sum, DownloadedAt: time.Now().UTC(), License: "CC BY 4.0"}
	stateFile, err := os.CreateTemp(filepath.Dir(path), ".geoip-state-*")
	if err != nil {
		return fallback("cannot create metadata file", err)
	}
	stateName := stateFile.Name()
	defer os.Remove(stateName)
	err = json.NewEncoder(stateFile).Encode(state)
	stateClose := stateFile.Close()
	if err != nil || stateClose != nil {
		return fallback("cannot write metadata", errors.New("metadata write failed"))
	}
	if err := os.Chmod(tempName, 0644); err != nil {
		return fallback("cannot set database permissions", err)
	}
	if err := os.Chmod(stateName, 0644); err != nil {
		return fallback("cannot set metadata permissions", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return fallback("cannot replace database; stop readers before updating on Windows", err)
	}
	// Each file is replaced separately; an interrupted sidecar update causes a fresh download next time.
	if err := os.Rename(stateName, path+".json"); err != nil {
		return Result{Status: "updated", Month: month, SHA256: sum, Warning: "database updated; metadata replacement failed"}, nil
	}
	return Result{Status: "updated", Month: month, SHA256: sum}, nil
}
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
