package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

func TestOIDCFlowAndAPIProtection(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test-key"))
	if err != nil {
		t.Fatal(err)
	}
	var issuerURL, nonce string
	issuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]any{"issuer": issuerURL, "authorization_endpoint": issuerURL + "/authorize", "token_endpoint": issuerURL + "/token", "jwks_uri": issuerURL + "/keys", "response_types_supported": []string{"code"}, "subject_types_supported": []string{"public"}, "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			if err := r.ParseForm(); err != nil || r.Form.Get("code_verifier") == "" {
				http.Error(w, "missing verifier", 400)
				return
			}
			claims, _ := json.Marshal(map[string]any{"iss": issuerURL, "aud": "test-client", "sub": "test-user", "exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix(), "nonce": nonce})
			signed, _ := signer.Sign(claims)
			token, _ := signed.CompactSerialize()
			json.NewEncoder(w).Encode(map[string]any{"access_token": "test-access", "token_type": "Bearer", "id_token": token})
		default:
			http.NotFound(w, r)
		}
	}))
	defer issuer.Close()
	issuerURL = issuer.URL
	cfg := testConfig()
	cfg.Issuer = issuer.URL
	cfg.ClientID = "test-client"
	cfg.ClientSecret = "test-secret"
	cfg.AllowHTTPOIDC = true
	cfg.SessionTTL = 15 * time.Minute
	app, err := New(context.Background(), cfg, fstest.MapFS{"index.html": {Data: []byte("PND")}})
	if err != nil {
		t.Fatal(err)
	}
	if w := requestApp(app, "POST", "/api/node-id", `{}`); w.Code != 401 {
		t.Fatal("API accessible without login")
	}
	login := requestApp(app, "GET", "/auth/login", "")
	dest, err := url.Parse(login.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if dest.Query().Get("code_challenge") == "" {
		t.Fatal("PKCE missing")
	}
	nonce = dest.Query().Get("nonce")
	if nonce == "" {
		t.Fatal("nonce missing")
	}
	flow := login.Result().Cookies()[0]
	bad := httptest.NewRequest("GET", "/auth/callback?state=wrong&code=test", nil)
	bad.AddCookie(flow)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, bad)
	if w.Code != 400 {
		t.Fatal("bad state accepted")
	}
	callback := httptest.NewRequest("GET", "/auth/callback?code=test&state="+url.QueryEscape(dest.Query().Get("state")), nil)
	callback.AddCookie(flow)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, callback)
	if w.Code != 303 {
		t.Fatal("callback failed", w.Code, w.Body.String())
	}
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "pnd_session" {
			cookie = c
		}
	}
	if cookie == nil || !cookie.HttpOnly {
		t.Fatal("session cookie missing")
	}
	if cookie.MaxAge != 900 {
		t.Fatal("configured session lifetime ignored")
	}
	check := httptest.NewRequest("GET", "/", nil)
	check.AddCookie(cookie)
	session, ok := app.auth.read(check, "pnd_session")
	if !ok || session.Expiry-session.IssuedAt != 900 {
		t.Fatal("signed lifetime differs from cookie")
	}
	req := httptest.NewRequest("POST", "/api/node-id", strings.NewReader("{}"))
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("authenticated API failed")
	}
	if len(w.Result().Cookies()) != 0 {
		t.Fatal("API activity renewed the session")
	}
	nonce = "wrong-nonce"
	w = httptest.NewRecorder()
	app.ServeHTTP(w, callback)
	if w.Code != 401 {
		t.Fatal("bad nonce accepted")
	}
}
