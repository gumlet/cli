package credentials

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const EnvKey = "GUMLET_API_KEY"

var ErrNoKey = errors.New("API key is not set — run 'gumlet login' or set " + EnvKey)

// Store holds the CLI's saved credential. The zero value uses ~/.gumlet.yaml.
type Store struct {
	Path string
}

type config struct {
	APIKey string `yaml:"api-key"`
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

func (s Store) APIKey() (string, error) {
	if key := strings.TrimSpace(os.Getenv(EnvKey)); key != "" {
		return key, nil
	}
	path, err := s.path()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNoKey
	}
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Parser errors can include the input, so don't echo them for secret files.
		return "", fmt.Errorf("invalid credential file %s; run 'gumlet login' to replace it", path)
	}
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return "", ErrNoKey
	}
	return key, nil
}

func (s Store) Save(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("API key must not be blank")
	}
	path, err := s.path()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(config{APIKey: key})
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
