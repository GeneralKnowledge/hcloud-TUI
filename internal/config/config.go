// Package config loads hcloud-tui configuration from the environment
// and an optional YAML config file.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrMissingToken indicates that no Hetzner Cloud API token was found.
var ErrMissingToken = errors.New("hetzner cloud API token not found")

// Config holds runtime configuration for hcloud-tui.
type Config struct {
	// Token is the Hetzner Cloud API token. Never log or display it in full.
	Token string
}

// Load resolves configuration with the following precedence (highest first):
//  1. HCLOUD_TOKEN environment variable
//  2. token field in the config file
//
// The optional configPath overrides the default config file location.
// An empty or missing config file is not an error.
func Load(configPath string) (Config, error) {
	cfg := Config{}

	path, err := resolveConfigPath(configPath)
	if err != nil {
		return Config{}, err
	}
	if path != "" {
		if _, statErr := os.Stat(path); statErr == nil {
			fileCfg, err := loadFile(path)
			if err != nil {
				return Config{}, err
			}
			cfg.Token = fileCfg.Token
		} else if !os.IsNotExist(statErr) {
			return Config{}, fmt.Errorf("stat config file %s: %w", path, statErr)
		} else if configPath != "" {
			// Explicit path was requested but missing — ignore and fall through
			// to environment variables / ErrMissingToken.
		}
	}

	if env := strings.TrimSpace(os.Getenv("HCLOUD_TOKEN")); env != "" {
		cfg.Token = env
	}

	if strings.TrimSpace(cfg.Token) == "" {
		return Config{}, ErrMissingToken
	}

	return cfg, nil
}

// MissingTokenMessage returns a user-friendly explanation of how to set a token.
func MissingTokenMessage() string {
	return `Hetzner Cloud API token not found.
Set the HCLOUD_TOKEN environment variable:
    export HCLOUD_TOKEN="..."
Then run hcloud-tui again.

Alternatively, place a token in ~/.config/hcloud-tui/config.yaml:
    token: "..."
Environment variables take precedence over the config file.`
}

func resolveConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	candidates := []string{
		filepath.Join(home, ".config", "hcloud-tui", "config.yaml"),
		filepath.Join(home, ".config", "hcloud-tui", "config.yml"),
		filepath.Join(home, ".hcloud-tui.yaml"),
		filepath.Join(home, ".hcloud-tui.yml"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", nil
}

type fileConfig struct {
	Token string
}

func loadFile(path string) (fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, fmt.Errorf("read config file %s: %w", path, err)
	}
	return parseSimpleYAML(string(data)), nil
}

// parseSimpleYAML extracts a top-level "token:" value without a YAML dependency.
// Supports:
//
//	token: value
//	token: "value"
//	token: 'value'
func parseSimpleYAML(content string) fileConfig {
	var cfg fileConfig
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) != "token" {
			continue
		}
		cfg.Token = unquote(strings.TrimSpace(value))
	}
	return cfg
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
