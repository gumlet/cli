package credentials

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const EnvKey = "GUMLET_API_KEY"

const refreshSkew = 60 * time.Second

var ErrNoKey = errors.New("not authenticated — run 'gumlet login' or set " + EnvKey)

// Token holds saved OAuth 2.0 credentials.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
}

// Refresher exchanges a refresh token for a fresh OAuth token.
type Refresher func(ctx context.Context, refreshToken string) (Token, error)

// Store holds the CLI's saved credentials. The zero value uses ~/.gumlet.yaml.
type Store struct {
	Path string
	Now  func() time.Time
}

type config struct {
	APIKey       string `yaml:"api-key,omitempty"`
	AccessToken  string `yaml:"access-token,omitempty"`
	RefreshToken string `yaml:"refresh-token,omitempty"`
	TokenType    string `yaml:"token-type,omitempty"`
	ExpiresAt    string `yaml:"expires-at,omitempty"`
}

func (s Store) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s Store) path() (string, error) {
	if s.Path != "" {
		return s.Path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, ".gumlet.yaml"), nil
}

func (s Store) readConfig() (config, string, error) {
	path, err := s.path()
	if err != nil {
		return config{}, "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return config{}, path, ErrNoKey
	}
	if err != nil {
		return config{}, path, fmt.Errorf("read credentials: %w", err)
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Parser errors can include the input, so don't echo them for secret files.
		return config{}, path, fmt.Errorf("invalid credential file %s; run 'gumlet login' to replace it", path)
	}
	return cfg, path, nil
}

// BearerToken resolves the bearer credential to send to the Gumlet API.
// GUMLET_API_KEY takes precedence for CI and non-interactive environments.
// Otherwise saved OAuth credentials are used and refreshed automatically if expired.
func (s Store) BearerToken(ctx context.Context, refresh Refresher) (string, error) {
	if key := strings.TrimSpace(os.Getenv(EnvKey)); key != "" {
		return key, nil
	}
	cfg, path, err := s.readConfig()
	if err != nil {
		return "", err
	}
	if accessToken := strings.TrimSpace(cfg.AccessToken); accessToken != "" {
		expiresAt, err := parseExpiresAt(cfg.ExpiresAt, path)
		if err != nil {
			return "", err
		}
		if expiresAt.IsZero() || s.now().Add(refreshSkew).Before(expiresAt) {
			return accessToken, nil
		}
		refreshToken := strings.TrimSpace(cfg.RefreshToken)
		if refreshToken == "" || refresh == nil {
			return "", errors.New("session expired — run 'gumlet login' or set " + EnvKey)
		}
		refreshed, err := refresh(ctx, refreshToken)
		if err != nil {
			return "", fmt.Errorf("session expired and token refresh failed (%w) — run 'gumlet login' or set %s", err, EnvKey)
		}
		refreshed.AccessToken = strings.TrimSpace(refreshed.AccessToken)
		if refreshed.AccessToken == "" {
			return "", errors.New("token refresh returned an empty access token — run 'gumlet login'")
		}
		if strings.TrimSpace(refreshed.RefreshToken) == "" {
			refreshed.RefreshToken = refreshToken
		}
		if strings.TrimSpace(refreshed.TokenType) == "" {
			refreshed.TokenType = strings.TrimSpace(cfg.TokenType)
		}
		_, _ = s.SaveToken(refreshed)
		return refreshed.AccessToken, nil
	}
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		return key, nil
	}
	return "", ErrNoKey
}

// LoadToken returns saved OAuth token fields from ~/.gumlet.yaml, or ErrNoKey if absent.
func (s Store) LoadToken() (Token, error) {
	cfg, path, err := s.readConfig()
	if err != nil {
		return Token{}, err
	}
	accessToken := strings.TrimSpace(cfg.AccessToken)
	refreshToken := strings.TrimSpace(cfg.RefreshToken)
	if accessToken == "" && refreshToken == "" {
		return Token{}, ErrNoKey
	}
	expiresAt, err := parseExpiresAt(cfg.ExpiresAt, path)
	if err != nil {
		return Token{}, err
	}
	return Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    strings.TrimSpace(cfg.TokenType),
		ExpiresAt:    expiresAt,
	}, nil
}

func parseExpiresAt(raw, path string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid credential file %s; run 'gumlet login' to replace it", path)
	}
	return parsed, nil
}

// SaveToken persists OAuth credentials to ~/.gumlet.yaml with 0600 permissions.
func (s Store) SaveToken(tok Token) (string, error) {
	tok.AccessToken = strings.TrimSpace(tok.AccessToken)
	if tok.AccessToken == "" {
		return "", errors.New("access token must not be blank")
	}
	cfg := config{
		AccessToken:  tok.AccessToken,
		RefreshToken: strings.TrimSpace(tok.RefreshToken),
		TokenType:    strings.TrimSpace(tok.TokenType),
	}
	if !tok.ExpiresAt.IsZero() {
		cfg.ExpiresAt = tok.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return s.writeConfig(cfg)
}

func (s Store) writeConfig(cfg config) (string, error) {
	path, err := s.path()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("encode credentials: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".gumlet-key-*")
	if err != nil {
		return "", fmt.Errorf("create credential file: %w", err)
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return "", fmt.Errorf("write credentials: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return "", fmt.Errorf("save credentials: %w", err)
	}
	return path, nil
}

func (s Store) Remove() (bool, error) {
	path, err := s.path()
	if err != nil {
		return false, err
	}
	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("remove credentials: %w", err)
	}
	return true, nil
}
