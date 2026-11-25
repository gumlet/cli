package credentials

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	t.Setenv(EnvKey, "")
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	store := Store{
		Path: filepath.Join(t.TempDir(), ".gumlet.yaml"),
		Now:  func() time.Time { return now },
	}

	if _, err := store.BearerToken(context.Background(), nil); !errors.Is(err, ErrNoKey) {
		t.Fatalf("missing file: %v", err)
	}
	if _, err := store.SaveToken(Token{AccessToken: " "}); err == nil {
		t.Fatal("saved a blank access token")
	}
	if err := os.WriteFile(store.Path, []byte("access-token: ["), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.BearerToken(context.Background(), nil); err == nil || errors.Is(err, ErrNoKey) {
		t.Fatalf("malformed file: %v", err)
	}

	exp := now.Add(time.Hour)
	if _, err := store.SaveToken(Token{
		AccessToken:  " fake-access ",
		RefreshToken: " fake-refresh ",
		TokenType:    "Bearer",
		ExpiresAt:    exp,
	}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf("credential permissions: %o", info.Mode().Perm())
	}

	if tok, err := store.BearerToken(context.Background(), nil); err != nil || tok != "fake-access" {
		t.Fatalf("saved access token: %q, %v", tok, err)
	}
	loaded, err := store.LoadToken()
	if err != nil || loaded.AccessToken != "fake-access" || loaded.RefreshToken != "fake-refresh" || !loaded.ExpiresAt.Equal(exp) {
		t.Fatalf("LoadToken: %+v, %v", loaded, err)
	}

	// Env override takes precedence over saved OAuth token.
	t.Setenv(EnvKey, " override ")
	if tok, err := store.BearerToken(context.Background(), nil); err != nil || tok != "override" {
		t.Fatalf("env override: %q, %v", tok, err)
	}
	t.Setenv(EnvKey, "")

	// Expired token refreshes and persists updated credentials.
	store.Now = func() time.Time { return exp.Add(time.Minute) }
	var refreshedWith string
	tok, err := store.BearerToken(context.Background(), func(_ context.Context, rt string) (Token, error) {
		refreshedWith = rt
		return Token{
			AccessToken: "fresh-access",
			ExpiresAt:   exp.Add(2 * time.Hour),
		}, nil
	})
	if err != nil || tok != "fresh-access" || refreshedWith != "fake-refresh" {
		t.Fatalf("refresh: tok=%q, rt=%q, err=%v", tok, refreshedWith, err)
	}
	// Existing refresh token should be preserved when omitted from refresh response.
	reloaded, err := store.LoadToken()
	if err != nil || reloaded.AccessToken != "fresh-access" || reloaded.RefreshToken != "fake-refresh" {
		t.Fatalf("reloaded after refresh: %+v, %v", reloaded, err)
	}

	// Expired token without refresh token returns session expired error.
	if _, err := store.SaveToken(Token{AccessToken: "only-access", ExpiresAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.BearerToken(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "session expired") {
		t.Fatalf("expected session expired error, got: %v", err)
	}

	// Legacy api-key file still works as fallback.
	if err := os.WriteFile(store.Path, []byte("api-key: legacy-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if tok, err := store.BearerToken(context.Background(), nil); err != nil || tok != "legacy-key" {
		t.Fatalf("legacy api-key: %q, %v", tok, err)
	}

	if removed, err := store.Remove(); !removed || err != nil {
		t.Fatalf("first remove: %t, %v", removed, err)
	}
	if removed, err := store.Remove(); removed || err != nil {
		t.Fatalf("second remove: %t, %v", removed, err)
	}
}
