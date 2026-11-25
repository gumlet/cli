package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginPKCEFlow(t *testing.T) {
	fixedNow := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

	var seenChallenge string
	var seenRedirectURI string

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.Form.Get("grant_type"); got != "authorization_code" {
			t.Errorf("grant_type = %q", got)
		}
		if got := r.Form.Get("client_id"); got != DefaultClientID {
			t.Errorf("client_id = %q", got)
		}
		if got := r.Form.Get("code"); got != "auth-code-123" {
			t.Errorf("code = %q", got)
		}
		if got := r.Form.Get("redirect_uri"); got != seenRedirectURI {
			t.Errorf("redirect_uri = %q; want %q", got, seenRedirectURI)
		}
		verifier := r.Form.Get("code_verifier")
		if pkceS256Challenge(verifier) != seenChallenge {
			t.Errorf("PKCE challenge mismatch: verifier=%q challenge=%q", verifier, seenChallenge)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "oauth-access-token",
			"refresh_token": "oauth-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer tokenSrv.Close()

	if got := (Config{}).WithDefaults().CallbackAddr; got != "127.0.0.1:8787" {
		t.Fatalf("default CallbackAddr = %q; want 127.0.0.1:8787", got)
	}

	cfg := Config{
		TokenURL:     tokenSrv.URL,
		CallbackAddr: "127.0.0.1:0",
		Now:          func() time.Time { return fixedNow },
		OpenBrowser: func(authURL string) error {
			u, err := url.Parse(authURL)
			if err != nil {
				t.Errorf("parse auth URL: %v", err)
				return err
			}
			q := u.Query()
			if q.Get("response_type") != "code" || q.Get("client_id") != DefaultClientID || q.Get("code_challenge_method") != "S256" {
				t.Errorf("unexpected auth URL query: %s", u.RawQuery)
			}
			if q.Has("scope") {
				t.Errorf("unexpected scope parameter when scopes empty: %s", u.RawQuery)
			}
			seenChallenge = q.Get("code_challenge")
			seenRedirectURI = q.Get("redirect_uri")
			state := q.Get("state")

			go func() {
				cbURL := seenRedirectURI + "?code=auth-code-123&state=" + url.QueryEscape(state)
				resp, err := http.Get(cbURL)
				if err != nil {
					t.Errorf("callback GET: %v", err)
					return
				}
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("callback status = %d; want 200", resp.StatusCode)
				}
			}()
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var notified bool
	tok, err := cfg.Login(ctx, LoginOptions{
		OnAuthURL: func(authURL string, opened bool, openErr error) {
			notified = true
			if !opened || openErr != nil || !strings.HasPrefix(authURL, DefaultAuthURL+"?") {
				t.Errorf("OnAuthURL(%q, %t, %v)", authURL, opened, openErr)
			}
		},
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if !notified {
		t.Error("OnAuthURL was not called")
	}
	if tok.AccessToken != "oauth-access-token" || tok.RefreshToken != "oauth-refresh-token" || tok.TokenType != "Bearer" {
		t.Fatalf("unexpected token: %+v", tok)
	}
	wantExp := fixedNow.Add(time.Hour)
	if !tok.ExpiresAt.Equal(wantExp) {
		t.Errorf("ExpiresAt = %v; want %v", tok.ExpiresAt, wantExp)
	}
}

func TestLoginStateMismatch(t *testing.T) {
	cfg := Config{
		CallbackAddr: "127.0.0.1:0",
		OpenBrowser: func(authURL string) error {
			u, _ := url.Parse(authURL)
			redirectURI := u.Query().Get("redirect_uri")
			go func() {
				resp, err := http.Get(redirectURI + "?code=abc&state=wrong-state")
				if err == nil {
					_ = resp.Body.Close()
				}
			}()
			return nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cfg.Login(ctx, LoginOptions{})
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("expected state mismatch error, got: %v", err)
	}
}

func TestLoginOAuthErrorCallback(t *testing.T) {
	cfg := Config{
		CallbackAddr: "127.0.0.1:0",
		OpenBrowser: func(authURL string) error {
			u, _ := url.Parse(authURL)
			redirectURI := u.Query().Get("redirect_uri")
			go func() {
				resp, err := http.Get(redirectURI + "?error=access_denied&error_description=user+cancelled")
				if err == nil {
					_ = resp.Body.Close()
				}
			}()
			return nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cfg.Login(ctx, LoginOptions{})
	if err == nil || !strings.Contains(err.Error(), "access_denied") {
		t.Fatalf("expected access_denied error, got: %v", err)
	}
}

func TestLoginBrowserOpenErrorNonFatal(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok"})
	}))
	defer tokenSrv.Close()

	cfg := Config{
		TokenURL:     tokenSrv.URL,
		CallbackAddr: "127.0.0.1:0",
		OpenBrowser: func(authURL string) error {
			u, _ := url.Parse(authURL)
			redirectURI := u.Query().Get("redirect_uri")
			state := u.Query().Get("state")
			go func() {
				resp, err := http.Get(redirectURI + "?code=ok&state=" + url.QueryEscape(state))
				if err == nil {
					_ = resp.Body.Close()
				}
			}()
			return errors.New("no display")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sawOpenErr bool
	tok, err := cfg.Login(ctx, LoginOptions{
		OnAuthURL: func(_ string, opened bool, openErr error) {
			sawOpenErr = !opened && openErr != nil
		},
	})
	if err != nil || tok.AccessToken != "tok" || !sawOpenErr {
		t.Fatalf("tok=%+v, err=%v, sawOpenErr=%t", tok, err, sawOpenErr)
	}
}

func TestRefreshAndRevoke(t *testing.T) {
	t.Run("refresh preserves existing refresh token when omitted", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "old-rt" || r.Form.Get("client_id") != DefaultClientID {
				t.Errorf("unexpected form: %v", r.Form)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "new-at",
				"expires_in":   1800,
			})
		}))
		defer srv.Close()

		tok, err := (Config{TokenURL: srv.URL}).Refresh(context.Background(), "old-rt")
		if err != nil {
			t.Fatalf("Refresh failed: %v", err)
		}
		if tok.AccessToken != "new-at" || tok.RefreshToken != "old-rt" {
			t.Fatalf("unexpected refreshed token: %+v", tok)
		}
	})

	t.Run("refresh rotates refresh token when returned", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-at",
				"refresh_token": "new-rt",
				"expires_in":    1800,
			})
		}))
		defer srv.Close()

		tok, err := (Config{TokenURL: srv.URL}).Refresh(context.Background(), "old-rt")
		if err != nil || tok.RefreshToken != "new-rt" {
			t.Fatalf("tok=%+v, err=%v", tok, err)
		}
	})

	t.Run("revoke posts token and hint", func(t *testing.T) {
		var called bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			_ = r.ParseForm()
			if r.Form.Get("client_id") != DefaultClientID || r.Form.Get("token") != "rt-1" || r.Form.Get("token_type_hint") != "refresh_token" {
				t.Errorf("unexpected revoke form: %v", r.Form)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		if err := (Config{RevokeURL: srv.URL}).Revoke(context.Background(), "rt-1", "refresh_token"); err != nil {
			t.Fatalf("Revoke failed: %v", err)
		}
		if !called {
			t.Fatal("revoke endpoint was not called")
		}
	})
}
