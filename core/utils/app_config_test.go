package utils

import (
	"strings"
	"testing"
)

func TestAppConfigSpotifyClientIDPrefersAuthClientID(t *testing.T) {
	cfg := AppConfig{}
	cfg.Auth.ClientID = "auth-client-id"

	if got := cfg.SpotifyClientID(); got != "auth-client-id" {
		t.Fatalf("SpotifyClientID() = %q, want %q", got, "auth-client-id")
	}
}

func TestAppConfigSpotifyClientIDTrimsWhitespace(t *testing.T) {
	cfg := AppConfig{}
	cfg.Auth.ClientID = " auth-client-id "

	if got := cfg.SpotifyClientID(); got != "auth-client-id" {
		t.Fatalf("SpotifyClientID() = %q, want %q", got, "auth-client-id")
	}
}

func TestValidateStartupConfigRequiresSpotifyClientID(t *testing.T) {
	err := validateStartupConfig(AppConfig{})
	if err == nil {
		t.Fatal("validateStartupConfig() returned nil, want error")
	}
	want := "missing required config value `auth.client_id`"
	if got := err.Error(); !strings.HasPrefix(got, want) {
		t.Fatalf("validateStartupConfig() error = %q, want prefix %q", got, want)
	}
}

func TestValidateStartupConfigAcceptsConfiguredSpotifyClientID(t *testing.T) {
	cfg := AppConfig{}
	cfg.Auth.ClientID = "configured-client-id"

	if err := validateStartupConfig(cfg); err != nil {
		t.Fatalf("validateStartupConfig() error = %v, want nil", err)
	}
}
