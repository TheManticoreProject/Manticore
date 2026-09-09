package server

import (
	"strings"
	"testing"
)

// TestDefaultIdentityIsWindowsPlausible guards the change away from "Unix" and
// "Manticore". These two strings are returned in every SESSION_SETUP_ANDX response,
// so a default-configured server described itself by naming this project — the most
// direct identification available to anyone observing a session.
func TestDefaultIdentityIsWindowsPlausible(t *testing.T) {
	// Non-empty is the original constraint: a strict client rejects a session setup
	// whose native strings are empty.
	if DefaultNativeOS == "" {
		t.Error("DefaultNativeOS is empty; strict clients reject the session setup")
	}
	if DefaultNativeLanMan == "" {
		t.Error("DefaultNativeLanMan is empty; strict clients reject the session setup")
	}

	// No default may name this project or another implementation.
	for _, banned := range []string{"Manticore", "manticore", "Samba", "Unix", "Go", "golang"} {
		for field, value := range map[string]string{
			"DefaultNativeOS":     DefaultNativeOS,
			"DefaultNativeLanMan": DefaultNativeLanMan,
		} {
			if strings.Contains(value, banned) {
				t.Errorf("%s = %q contains %q", field, value, banned)
			}
		}
	}

	// The pair must describe the same product, the way Windows does: one field
	// carries the build number and the other the major.minor version.
	if !strings.HasPrefix(DefaultNativeOS, "Windows ") {
		t.Errorf("DefaultNativeOS = %q, want a Windows product string", DefaultNativeOS)
	}
	if !strings.HasPrefix(DefaultNativeLanMan, "Windows ") {
		t.Errorf("DefaultNativeLanMan = %q, want a Windows product string", DefaultNativeLanMan)
	}
}

// TestNewServerAppliesIdentityDefaults checks the defaults actually reach the
// configuration, and that an operator-supplied value is preserved.
func TestNewServerAppliesIdentityDefaults(t *testing.T) {
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if got := srv.config.NativeOS; got != DefaultNativeOS {
		t.Errorf("NativeOS = %q, want the default %q", got, DefaultNativeOS)
	}
	if got := srv.config.NativeLanMan; got != DefaultNativeLanMan {
		t.Errorf("NativeLanMan = %q, want the default %q", got, DefaultNativeLanMan)
	}

	custom, err := NewServer(Config{NativeOS: "Windows Server 2019 Standard 17763", NativeLanMan: "Windows Server 2019 Standard 10.0"})
	if err != nil {
		t.Fatalf("NewServer with overrides: %v", err)
	}
	if got := custom.config.NativeOS; got != "Windows Server 2019 Standard 17763" {
		t.Errorf("operator NativeOS was overwritten: got %q", got)
	}
}
