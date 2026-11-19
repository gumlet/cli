package credentials

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStore(t *testing.T) {
	t.Setenv(EnvKey, "")
	store := Store{Path: filepath.Join(t.TempDir(), ".gumlet.yaml")}
	if _, err := store.APIKey(); !errors.Is(err, ErrNoKey) {
		t.Fatalf("missing file: %v", err)
	}
	if _, err := store.Save(" "); err == nil {
		t.Fatal("saved a blank key")
	}
	if err := os.WriteFile(store.Path, []byte("api-key: ["), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.APIKey(); err == nil || errors.Is(err, ErrNoKey) {
		t.Fatalf("malformed file: %v", err)
	}
	if _, err := store.Save(" fake-key "); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf("credential permissions: %o", info.Mode().Perm())
	}
	if key, err := store.APIKey(); err != nil || key != "fake-key" {
		t.Fatalf("saved key: %q, %v", key, err)
	}
	t.Setenv(EnvKey, " override ")
	if key, err := store.APIKey(); err != nil || key != "override" {
		t.Fatalf("env override: %q, %v", key, err)
	}
	if removed, err := store.Remove(); !removed || err != nil {
		t.Fatalf("first remove: %t, %v", removed, err)
	}
	if removed, err := store.Remove(); removed || err != nil {
		t.Fatalf("second remove: %t, %v", removed, err)
	}
}
