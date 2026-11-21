package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionDurationParsing(t *testing.T) {
	for _, value := range []string{"1m", "15m", "30m", "1h", "24h", "90s"} {
		if _, err := sessionDuration(value); err != nil {
			t.Fatalf("rejected %s", value)
		}
	}
	for _, value := range []string{"", "0", "0m", "-1m", "59s", "25h", "forever", "90.5s"} {
		if _, err := sessionDuration(value); err == nil {
			t.Fatalf("accepted %s", value)
		}
	}
}

func TestStatelessAbsoluteSessionExpiry(t *testing.T) {
	start := time.Unix(2000000000, 0)
	now := start
	a := &auth{signing: []byte("test-only-signing-key"), sessionTTL: 30 * time.Minute, clock: func() time.Time { return now }}
	signed := a.sign(ticket{IssuedAt: start.Unix(), Expiry: start.Add(30 * time.Minute).Unix(), Session: true})
	request := func(value string) *http.Request {
		r := httptest.NewRequest("GET", "/api/meta", nil)
		r.AddCookie(&http.Cookie{Name: "pnd_session", Value: value})
		return r
	}
	r := request(signed)
	now = start.Add(29*time.Minute + 59*time.Second)
	if !a.logged(r) {
		t.Fatal("valid session rejected")
	}
	recreated := &auth{signing: a.signing, sessionTTL: a.sessionTTL, clock: a.clock}
	if !recreated.logged(r) {
		t.Fatal("session depends on instance state")
	}
	now = start.Add(30 * time.Minute)
	if a.logged(r) {
		t.Fatal("session valid at expiry")
	}
	now = start.Add(10 * time.Minute)
	a.sessionTTL = 5 * time.Minute
	if a.logged(r) {
		t.Fatal("shorter configured TTL was ignored")
	}
	a.sessionTTL = time.Hour
	now = start.Add(31 * time.Minute)
	if a.logged(r) {
		t.Fatal("larger TTL extended signed expiry")
	}
	now = start
	if a.logged(request(a.sign(ticket{Expiry: start.Add(8 * time.Hour).Unix(), Session: true}))) {
		t.Fatal("legacy cookie without issue time accepted")
	}
	if a.logged(request(a.sign(ticket{IssuedAt: start.Add(time.Minute).Unix(), Expiry: start.Add(time.Hour).Unix(), Session: true}))) {
		t.Fatal("future issue time accepted")
	}
}

func TestCookieExpiryMatchesConfiguredLifetime(t *testing.T) {
	now := time.Unix(2000000000, 0)
	a := &auth{clock: func() time.Time { return now }, secure: true}
	w := httptest.NewRecorder()
	a.cookie(w, "pnd_session", "test-value", 900)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("cookie missing")
	}
	c := cookies[0]
	if c.MaxAge != 900 || !c.Expires.Equal(now.Add(15*time.Minute)) || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode {
		t.Fatal("cookie lifetime or attributes wrong")
	}
}
