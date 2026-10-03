package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromEnvTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("token = file-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom("env-token", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Fatalf("token = %q, want env-token", cfg.Token)
	}
	if cfg.TokenSource != "env" {
		t.Fatalf("TokenSource = %q, want env", cfg.TokenSource)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("# comment\ntoken = \"file-token\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom("", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "file-token" {
		t.Fatalf("token = %q, want file-token", cfg.Token)
	}
	if cfg.TokenSource != "config" {
		t.Fatalf("TokenSource = %q, want config", cfg.TokenSource)
	}
}

func TestLoadFromMissingToken(t *testing.T) {
	_, err := LoadFrom("", filepath.Join(t.TempDir(), "missing"))
	if err != ErrTokenNotFound {
		t.Fatalf("error = %v, want ErrTokenNotFound", err)
	}
}

func TestLoadFromEmptyEnvFallsBackToConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("hcloud_token=legacy-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom("   ", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "legacy-token" {
		t.Fatalf("token = %q, want legacy-token", cfg.Token)
	}
}

func TestTokenMissingMessage(t *testing.T) {
	msg := TokenMissingMessage()
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	if !strings.Contains(msg, "HCLOUD_TOKEN") {
		t.Fatalf("message should mention HCLOUD_TOKEN: %s", msg)
	}
}
