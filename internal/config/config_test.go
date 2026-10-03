package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GeneralKnowledge/hcloud-TUI/internal/config"
)

func TestLoad_EnvToken(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "env-token-value")

	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Token != "env-token-value" {
		t.Fatalf("Token = %q, want env-token-value", cfg.Token)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "")

	_, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("Load() error = nil, want ErrMissingToken")
	}
	if !isMissingToken(err) {
		t.Fatalf("Load() error = %v, want ErrMissingToken", err)
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("token: file-token-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Token != "file-token-value" {
		t.Fatalf("Token = %q, want file-token-value", cfg.Token)
	}
}

func TestLoad_EnvTakesPrecedence(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "env-wins")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("token: \"file-token\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Token != "env-wins" {
		t.Fatalf("Token = %q, want env-wins", cfg.Token)
	}
}

func TestLoad_QuotedConfigToken(t *testing.T) {
	t.Setenv("HCLOUD_TOKEN", "")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("token: 'quoted-token'\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Token != "quoted-token" {
		t.Fatalf("Token = %q, want quoted-token", cfg.Token)
	}
}

func TestMissingTokenMessage(t *testing.T) {
	msg := config.MissingTokenMessage()
	if msg == "" {
		t.Fatal("MissingTokenMessage() empty")
	}
	if !contains(msg, "HCLOUD_TOKEN") {
		t.Fatalf("message missing HCLOUD_TOKEN: %s", msg)
	}
}

func isMissingToken(err error) bool {
	return err == config.ErrMissingToken
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i+len(substr) <= len(s); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
