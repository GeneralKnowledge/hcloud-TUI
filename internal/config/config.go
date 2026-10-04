// Package config loads hcloud-tui configuration from the environment and an optional config file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvToken is the primary authentication environment variable.
	EnvToken = "HCLOUD_TOKEN"
	// DefaultConfigRelPath is the default config file relative to the user's home directory.
	DefaultConfigRelPath = ".config/hcloud-tui/config"
)

// ErrTokenNotFound indicates that no Hetzner Cloud API token was configured.
var ErrTokenNotFound = errors.New("hetzner cloud api token not found")

// Config holds runtime configuration for hcloud-tui.
type Config struct {
	Token      string
	TokenSource string // "env", "config", or ""
	ConfigPath string
}

// Load resolves configuration with environment variables taking precedence over the config file.
func Load() (Config, error) {
	return LoadFrom(os.Getenv(EnvToken), defaultConfigPath())
}

// LoadFrom resolves configuration from an explicit token and config file path.
// An empty tokenEnv means the environment variable was unset.
// An empty configPath skips the config file.
func LoadFrom(tokenEnv, configPath string) (Config, error) {
	cfg := Config{ConfigPath: configPath}

	tokenEnv = strings.TrimSpace(tokenEnv)
	if tokenEnv != "" {
		cfg.Token = tokenEnv
		cfg.TokenSource = "env"
		return cfg, nil
	}

	if configPath != "" {
		token, err := readTokenFromFile(configPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
		if token != "" {
			cfg.Token = token
			cfg.TokenSource = "config"
			return cfg, nil
		}
	}

	return Config{}, ErrTokenNotFound
}

// TokenMissingMessage returns a user-facing explanation for a missing token.
func TokenMissingMessage() string {
	return `Hetzner Cloud API token not found.
Set the HCLOUD_TOKEN environment variable:
    export HCLOUD_TOKEN="..."
Then run hcloud-tui again.

Alternatively, place the token in:
    ~/.config/hcloud-tui/config
as:
    token = "..."`
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, DefaultConfigRelPath)
}

func readTokenFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "token" || key == "hcloud_token" {
			if value == "" {
				return "", fmt.Errorf("config file %s has an empty token", path)
			}
			return value, nil
		}
	}
	return "", nil
}
