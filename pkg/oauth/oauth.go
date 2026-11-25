package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	DefaultAuthURL      = "https://webapp.gumlet.com/oauth/authorize"
	DefaultTokenURL     = "https://webapp.gumlet.com/oauth/token"
	DefaultRevokeURL    = "https://webapp.gumlet.com/oauth/revoke"
	DefaultClientID     = "97b2fdf8-ba33-42c3-bf58-a17233d092b5"
	DefaultCallbackAddr = "127.0.0.1:8787"

	maxResponseBytes = 1 << 20 // 1 MiB
)

// Config configures the OAuth 2.0 PKCE client. Zero fields fall back to Gumlet defaults.
type Config struct {
	AuthURL      string
	TokenURL     string
	RevokeURL    string
	ClientID     string
	CallbackAddr string
	Scopes       []string
	HTTPClient   *http.Client
	OpenBrowser  func(string) error
	Now          func() time.Time
}

// Token holds the OAuth 2.0 tokens returned by the token endpoint.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
}

// LoginOptions customizes an interactive browser login.
type LoginOptions struct {
	NoBrowser bool
	// OnAuthURL is called with the authorization URL and whether opening the
	// browser will be attempted.
	OnAuthURL func(authURL string, openedBrowser bool, openErr error)
}

// WithDefaults returns a copy of Config with zero fields populated with Gumlet defaults.
func (c Config) WithDefaults() Config {
	if c.AuthURL == "" {
		c.AuthURL = DefaultAuthURL
	}
	if c.TokenURL == "" {
		c.TokenURL = DefaultTokenURL
	}
	if c.RevokeURL == "" {
		c.RevokeURL = DefaultRevokeURL
	}
	if c.ClientID == "" {
		c.ClientID = DefaultClientID
	}
	if c.CallbackAddr == "" {
		c.CallbackAddr = DefaultCallbackAddr
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.OpenBrowser == nil {
		c.OpenBrowser = openBrowser
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	return c
}

// Login executes the OAuth 2.0 Authorization Code flow with PKCE (S256) using
// a localhost callback server (defaulting to http://127.0.0.1:8787/callback).
func (c Config) Login(ctx context.Context, opts LoginOptions) (Token, error) {
	c = c.WithDefaults()

	verifier, err := randomBase64URL(32)
	if err != nil {
		return Token{}, fmt.Errorf("generate PKCE verifier: %w", err)
	}
	challenge := pkceS256Challenge(verifier)

	state, err := randomBase64URL(32)
	if err != nil {
		return Token{}, fmt.Errorf("generate OAuth state: %w", err)
	}

	listener, err := net.Listen("tcp", c.CallbackAddr)
	if err != nil {
		return Token{}, fmt.Errorf("start local callback server on %s: %w", c.CallbackAddr, err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	authURL, err := c.buildAuthURL(redirectURI, state, challenge)
	if err != nil {
		return Token{}, err
	}

	type callbackResult struct {
		code string
		err  error
	}
	resultCh := make(chan callbackResult, 1)
	var once sync.Once
	deliver := func(res callbackResult) {
		once.Do(func() {
			resultCh <- res
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		q := r.URL.Query()
		if oauthErr := q.Get("error"); oauthErr != "" {
			desc := q.Get("error_description")
			msg := oauthErr
			if desc != "" {
				msg = fmt.Sprintf("%s: %s", oauthErr, desc)
			}
			writeCallbackHTML(w, http.StatusBadRequest, "Authentication failed", msg)
			deliver(callbackResult{err: fmt.Errorf("authorization failed: %s", msg)})
			return
		}
		gotState := q.Get("state")
		if len(gotState) == 0 || subtle.ConstantTimeCompare([]byte(gotState), []byte(state)) != 1 {
			writeCallbackHTML(w, http.StatusBadRequest, "Authentication failed", "Invalid state parameter.")
			deliver(callbackResult{err: errors.New("authorization failed: state mismatch")})
			return
		}
		code := q.Get("code")
		if code == "" {
			writeCallbackHTML(w, http.StatusBadRequest, "Authentication failed", "Missing authorization code.")
			deliver(callbackResult{err: errors.New("authorization failed: missing authorization code")})
			return
		}
		writeCallbackHTML(w, http.StatusOK, "Authentication complete", "You have signed in to Gumlet CLI. You can close this tab and return to your terminal.")
		deliver(callbackResult{code: code})
	})

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			deliver(callbackResult{err: fmt.Errorf("callback server failed: %w", serveErr)})
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	var openErr error
	opened := false
	if !opts.NoBrowser {
		openErr = c.OpenBrowser(authURL)
		opened = openErr == nil
	}
	if opts.OnAuthURL != nil {
		opts.OnAuthURL(authURL, opened, openErr)
	}

	var code string
	select {
	case <-ctx.Done():
		return Token{}, fmt.Errorf("login cancelled or timed out: %w", ctx.Err())
	case res := <-resultCh:
		if res.err != nil {
			return Token{}, res.err
		}
		code = res.code
	}

	return c.ExchangeCode(ctx, code, redirectURI, verifier)
}

// ExchangeCode exchanges an authorization code and PKCE verifier for tokens.
func (c Config) ExchangeCode(ctx context.Context, code, redirectURI, verifier string) (Token, error) {
	c = c.WithDefaults()
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {c.ClientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
	return c.postToken(ctx, form)
}

// Refresh exchanges a refresh token for a fresh access token.
func (c Config) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	c = c.WithDefaults()
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return Token{}, errors.New("refresh token is empty")
	}
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.ClientID},
		"refresh_token": {refreshToken},
	}
	tok, err := c.postToken(ctx, form)
	if err != nil {
		return Token{}, err
	}
	if tok.RefreshToken == "" {
		tok.RefreshToken = refreshToken
	}
	return tok, nil
}

// Revoke revokes an access or refresh token at the OAuth revocation endpoint (RFC 7009).
func (c Config) Revoke(ctx context.Context, token, tokenTypeHint string) error {
	c = c.WithDefaults()
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	form := url.Values{
		"client_id": {c.ClientID},
		"token":     {token},
	}
	if tokenTypeHint != "" {
		form.Set("token_type_hint", tokenTypeHint)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.RevokeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseOAuthError("revoke token", resp.StatusCode, body)
	}
	return nil
}

func (c Config) buildAuthURL(redirectURI, state, challenge string) (string, error) {
	u, err := url.Parse(c.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %w", err)
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if len(c.Scopes) > 0 {
		q.Set("scope", strings.Join(c.Scopes, " "))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c Config) postToken(ctx context.Context, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return Token{}, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Token{}, parseOAuthError("token exchange", resp.StatusCode, body)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return Token{}, fmt.Errorf("decode token response: %w", err)
	}
	if tr.Error != "" {
		if tr.ErrorDescription != "" {
			return Token{}, fmt.Errorf("token exchange: %s: %s", tr.Error, tr.ErrorDescription)
		}
		return Token{}, fmt.Errorf("token exchange: %s", tr.Error)
	}
	tr.AccessToken = strings.TrimSpace(tr.AccessToken)
	if tr.AccessToken == "" {
		return Token{}, errors.New("token response missing access_token")
	}

	tok := Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: strings.TrimSpace(tr.RefreshToken),
		TokenType:    strings.TrimSpace(tr.TokenType),
	}
	if tr.ExpiresIn > 0 {
		tok.ExpiresAt = c.Now().UTC().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	return tok, nil
}

func parseOAuthError(action string, status int, body []byte) error {
	var errResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		if errResp.ErrorDescription != "" {
			return fmt.Errorf("%s failed (HTTP %d): %s: %s", action, status, errResp.Error, errResp.ErrorDescription)
		}
		return fmt.Errorf("%s failed (HTTP %d): %s", action, status, errResp.Error)
	}
	return fmt.Errorf("%s failed (HTTP %d)", action, status)
}

func randomBase64URL(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func pkceS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func writeCallbackHTML(w http.ResponseWriter, status int, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>%s — Gumlet CLI</title></head>
<body style="font-family:system-ui,-apple-system,sans-serif;max-width:480px;margin:80px auto;padding:0 20px;color:#111">
<h2>%s</h2>
<p>%s</p>
</body>
</html>`, html.EscapeString(title), html.EscapeString(title), html.EscapeString(message))
}

func openBrowser(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}
