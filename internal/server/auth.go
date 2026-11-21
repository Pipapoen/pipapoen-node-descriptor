package server

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type auth struct {
	sessionTTL time.Duration
	clock      func() time.Time
	oauth      oauth2.Config
	verifier   *oidc.IDTokenVerifier
	signing    []byte
	secure     bool
	client     *http.Client
}
type ticket struct {
	IssuedAt int64  `json:"issuedAt,omitempty"`
	Expiry   int64  `json:"expiry"`
	State    string `json:"state,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
	Verifier string `json:"verifier,omitempty"`
	Session  bool   `json:"session,omitempty"`
}

func newAuth(ctx context.Context, c Config) (*auth, error) {
	if c.Issuer == "" {
		return nil, nil
	}
	client := &http.Client{Timeout: 12 * time.Second}
	ctx = oidc.ClientContext(ctx, client)
	p, err := oidc.NewProvider(ctx, c.Issuer)
	if err != nil {
		return nil, errors.New("OIDC discovery failed")
	}
	mac := hmac.New(sha256.New, c.Key)
	mac.Write([]byte("PND/session-signing/v1"))
	return &auth{oauth: oauth2.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, Endpoint: p.Endpoint(), RedirectURL: c.PublicURL + "/auth/callback", Scopes: []string{oidc.ScopeOpenID}}, verifier: p.Verifier(&oidc.Config{ClientID: c.ClientID}), signing: mac.Sum(nil), secure: strings.HasPrefix(c.PublicURL, "https://"), client: client, sessionTTL: c.SessionTTL, clock: time.Now}, nil
}
func (a *auth) now() time.Time {
	if a.clock != nil {
		return a.clock()
	}
	return time.Now()
}
func randomString() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func (a *auth) sign(v ticket) string {
	b, _ := json.Marshal(v)
	s := base64.RawURLEncoding.EncodeToString(b)
	m := hmac.New(sha256.New, a.signing)
	m.Write([]byte(s))
	return s + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func (a *auth) read(r *http.Request, name string) (ticket, bool) {
	var v ticket
	c, e := r.Cookie(name)
	if e != nil || len(c.Value) > 4096 {
		return v, false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return v, false
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return v, false
	}
	m := hmac.New(sha256.New, a.signing)
	m.Write([]byte(parts[0]))
	if !hmac.Equal(sig, m.Sum(nil)) {
		return v, false
	}
	b, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil || json.Unmarshal(b, &v) != nil || v.Expiry <= a.now().Unix() {
		return v, false
	}
	return v, true
}
func (a *auth) cookie(w http.ResponseWriter, name, value string, age int) {
	c := &http.Cookie{Name: name, Value: value, Path: "/", HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteLaxMode, MaxAge: age}
	if age > 0 {
		c.Expires = a.now().Add(time.Duration(age) * time.Second)
	}
	http.SetCookie(w, c)
}
func (a *auth) logged(r *http.Request) bool {
	if a == nil {
		return true
	}
	t, ok := a.read(r, "pnd_session")
	now := a.now().Unix()
	return ok && t.Session && t.IssuedAt > 0 && t.IssuedAt <= now && t.Expiry > t.IssuedAt && now-t.IssuedAt < int64(a.sessionTTL/time.Second)
}
func (a *auth) login(w http.ResponseWriter, r *http.Request) {
	t := ticket{Expiry: a.now().Add(5 * time.Minute).Unix(), State: randomString(), Nonce: randomString(), Verifier: oauth2.GenerateVerifier()}
	a.cookie(w, "pnd_flow", a.sign(t), 300)
	http.Redirect(w, r, a.oauth.AuthCodeURL(t.State, oidc.Nonce(t.Nonce), oauth2.S256ChallengeOption(t.Verifier)), http.StatusFound)
}
func (a *auth) callback(w http.ResponseWriter, r *http.Request) {
	t, ok := a.read(r, "pnd_flow")
	a.cookie(w, "pnd_flow", "", -1)
	if !ok || t.State == "" || subtle.ConstantTimeCompare([]byte(t.State), []byte(r.URL.Query().Get("state"))) != 1 || r.URL.Query().Get("code") == "" {
		http.Redirect(w, r, "/?auth=error", http.StatusSeeOther)
		return
	}
	ctx := oidc.ClientContext(r.Context(), a.client)
	tok, err := a.oauth.Exchange(ctx, r.URL.Query().Get("code"), oauth2.VerifierOption(t.Verifier))
	if err != nil {
		http.Redirect(w, r, "/?auth=error", http.StatusSeeOther)
		return
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok {
		http.Redirect(w, r, "/?auth=error", http.StatusSeeOther)
		return
	}
	id, err := a.verifier.Verify(ctx, raw)
	if err != nil || subtle.ConstantTimeCompare([]byte(id.Nonce), []byte(t.Nonce)) != 1 {
		http.Redirect(w, r, "/?auth=error", http.StatusSeeOther)
		return
	}
	now := a.now()
	a.cookie(w, "pnd_session", a.sign(ticket{IssuedAt: now.Unix(), Expiry: now.Add(a.sessionTTL).Unix(), Session: true, Nonce: randomString()}), int(a.sessionTTL/time.Second))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
