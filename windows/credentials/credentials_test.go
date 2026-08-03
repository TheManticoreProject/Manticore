package credentials_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestIsDomainIdentity(t *testing.T) {
	creds := credentials.Credentials{Domain: "example.com"}
	if !creds.IsDomainIdentity() {
		t.Errorf("Expected true, got false")
	}

	creds = credentials.Credentials{Domain: ""}
	if creds.IsDomainIdentity() {
		t.Errorf("Expected false, got true")
	}
}

func TestIsLocalIdentity(t *testing.T) {
	creds := credentials.Credentials{Domain: ""}
	if !creds.IsLocalIdentity() {
		t.Errorf("Expected true, got false")
	}

	creds = credentials.Credentials{Domain: "example.com"}
	if creds.IsLocalIdentity() {
		t.Errorf("Expected false, got true")
	}
}

func TestCanPassTheHash(t *testing.T) {
	creds := credentials.Credentials{NTHash: "31d6cfe0d16ae931b73c59d7e0c089c0", Username: "user"}
	if !creds.CanPassTheHash() {
		t.Errorf("Expected true, got false")
	}

	creds = credentials.Credentials{NTHash: "", Username: "user"}
	if creds.CanPassTheHash() {
		t.Errorf("Expected false, got true")
	}

	creds = credentials.Credentials{NTHash: "31d6cfe0d16ae931b73c59d7e0c089c0", Username: ""}
	if creds.CanPassTheHash() {
		t.Errorf("Expected false, got true")
	}
}

func TestGetLMHash(t *testing.T) {
	creds := credentials.Credentials{LMHash: "aad3b435b51404eeaad3b435b51404ee"}
	if creds.GetLMHash() != "aad3b435b51404eeaad3b435b51404ee" {
		t.Errorf("Expected lmhash, got %s", creds.GetLMHash())
	}
}

func TestGetNTHash(t *testing.T) {
	creds := credentials.Credentials{NTHash: "31d6cfe0d16ae931b73c59d7e0c089c0"}
	if creds.GetNTHash() != "31d6cfe0d16ae931b73c59d7e0c089c0" {
		t.Errorf("Expected nthash, got %s", creds.GetNTHash())
	}
}

func TestGetDomain(t *testing.T) {
	creds := credentials.Credentials{Domain: "example.com"}
	if creds.GetDomain() != "example.com" {
		t.Errorf("Expected example.com, got %s", creds.GetDomain())
	}
}

func TestGetUsername(t *testing.T) {
	creds := credentials.Credentials{Username: "user"}
	if creds.GetUsername() != "user" {
		t.Errorf("Expected user, got %s", creds.GetUsername())
	}
}

func TestGetPassword(t *testing.T) {
	creds := credentials.Credentials{Password: "password"}
	if creds.GetPassword() != "password" {
		t.Errorf("Expected password, got %s", creds.GetPassword())
	}
}

func TestParseLMNTHashes(t *testing.T) {
	testCases := []struct {
		name        string
		authHashes  string
		expectedLM  string
		expectedNT  string
		expectError bool
	}{
		{name: "Valid Hashes", authHashes: "aad3b435b51404eeaad3b435b51404ee:31d6cfe0d16ae931b73c59d7e0c089c0", expectedLM: "aad3b435b51404eeaad3b435b51404ee", expectedNT: "31d6cfe0d16ae931b73c59d7e0c089c0", expectError: false},
		{name: "Valid NT Hash Only", authHashes: "31d6cfe0d16ae931b73c59d7e0c089c0", expectedLM: "", expectedNT: "31d6cfe0d16ae931b73c59d7e0c089c0", expectError: false},
		{name: "Invalid Hash Format", authHashes: "invalidhash", expectedLM: "", expectedNT: "", expectError: true},
		{name: "Invalid LM Hash Length", authHashes: "aad3b435b51404eeaad3b435b51404:31d6cfe0d16ae931b73c59d7e0c089c0", expectedLM: "", expectedNT: "", expectError: true},
		{name: "Invalid NT Hash Length", authHashes: "aad3b435b51404eeaad3b435b51404ee:31d6cfe0d16ae931b73c59d7e0c089", expectedLM: "", expectedNT: "", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			lmHash, ntHash, err := credentials.ParseLMNTHashes(testCase.authHashes)
			if testCase.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected nil, got error: %s", err)
				}
				if lmHash != testCase.expectedLM {
					t.Errorf("Expected LM hash %s, got %s", testCase.expectedLM, lmHash)
				}
				if ntHash != testCase.expectedNT {
					t.Errorf("Expected NT hash %s, got %s", testCase.expectedNT, ntHash)
				}
			}
		})
	}
}

func TestSetAESKey(t *testing.T) {
	// 32 hex characters is a 16-byte aes128 key, 64 a 32-byte aes256 key.
	aes128 := strings.Repeat("ab", 16)
	aes256 := strings.Repeat("cd", 32)

	tests := []struct {
		name    string
		hexKey  string
		wantErr bool
		want    string
	}{
		{name: "aes128", hexKey: aes128, want: aes128},
		{name: "aes256", hexKey: aes256, want: aes256},
		{name: "surrounding whitespace is trimmed", hexKey: "  " + aes256 + "\n", want: aes256},
		{name: "uppercase hex", hexKey: strings.ToUpper(aes256), want: strings.ToUpper(aes256)},
		{name: "not hex", hexKey: strings.Repeat("zz", 32), wantErr: true},
		{name: "odd length", hexKey: strings.Repeat("a", 63), wantErr: true},
		{name: "too short", hexKey: strings.Repeat("ab", 8), wantErr: true},
		{name: "too long", hexKey: strings.Repeat("ab", 48), wantErr: true},
		{name: "empty", hexKey: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &credentials.Credentials{Username: "user"}
			err := c.SetAESKey(tt.hexKey)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("SetAESKey(%q) error = nil, want an error", tt.hexKey)
				}
				if c.GetAESKey() != "" {
					t.Errorf("AESKey = %q after a rejected key, want it left empty", c.GetAESKey())
				}
				if c.CanUseAESKey() {
					t.Error("CanUseAESKey() = true after a rejected key")
				}
				return
			}

			if err != nil {
				t.Fatalf("SetAESKey(%q) error = %v", tt.hexKey, err)
			}
			if got := c.GetAESKey(); got != tt.want {
				t.Errorf("GetAESKey() = %q, want %q", got, tt.want)
			}
			if !c.CanUseAESKey() {
				t.Error("CanUseAESKey() = false, want true")
			}
		})
	}
}

func TestCanUseAESKeyRequiresUsername(t *testing.T) {
	c := &credentials.Credentials{}
	if err := c.SetAESKey(strings.Repeat("ab", 32)); err != nil {
		t.Fatalf("SetAESKey() error = %v", err)
	}
	if c.CanUseAESKey() {
		t.Error("CanUseAESKey() = true with no username, want false")
	}
}

func TestSetKeytab(t *testing.T) {
	path := filepath.Join(t.TempDir(), "user.keytab")
	if err := os.WriteFile(path, []byte("not a real keytab"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	c := &credentials.Credentials{Username: "user"}
	if err := c.SetKeytab(path); err != nil {
		t.Fatalf("SetKeytab(%q) error = %v", path, err)
	}
	if got := c.GetKeytab(); got != path {
		t.Errorf("GetKeytab() = %q, want %q", got, path)
	}
	if !c.CanUseKeytab() {
		t.Error("CanUseKeytab() = false, want true")
	}
}

func TestSetKeytabRejectsMissingFile(t *testing.T) {
	c := &credentials.Credentials{Username: "user"}
	missing := filepath.Join(t.TempDir(), "absent.keytab")

	if err := c.SetKeytab(missing); err == nil {
		t.Fatal("SetKeytab() with a missing file error = nil, want an error")
	}
	if c.GetKeytab() != "" {
		t.Errorf("Keytab = %q after a rejected path, want it left empty", c.GetKeytab())
	}
	if c.CanUseKeytab() {
		t.Error("CanUseKeytab() = true after a rejected path")
	}
}

func TestSetCCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "krb5cc")
	if err := os.WriteFile(path, []byte("ticket cache bytes"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	// The principal comes from the ticket, so no username is required.
	c := &credentials.Credentials{}
	if c.CanUseCCache() {
		t.Error("CanUseCCache() = true before a ccache is set")
	}
	if err := c.SetCCache(path); err != nil {
		t.Fatalf("SetCCache(%q) error = %v", path, err)
	}
	if got := c.GetCCache(); got != path {
		t.Errorf("GetCCache() = %q, want %q", got, path)
	}
	if !c.CanUseCCache() {
		t.Error("CanUseCCache() = false after a readable ccache is set")
	}
}

func TestSetCCacheRejectsMissingFile(t *testing.T) {
	c := &credentials.Credentials{}
	if err := c.SetCCache(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("SetCCache() with a missing file error = nil, want an error")
	}
	if c.GetCCache() != "" {
		t.Errorf("CCache = %q after a rejected path, want it left empty", c.GetCCache())
	}
	if c.CanUseCCache() {
		t.Error("CanUseCCache() = true after a rejected path")
	}
}

func TestSetKirbi(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tgt.kirbi")
	if err := os.WriteFile(path, []byte("kirbi bytes"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	c := &credentials.Credentials{}
	if err := c.SetKirbi(path); err != nil {
		t.Fatalf("SetKirbi(%q) error = %v", path, err)
	}
	if got := c.GetKirbi(); got != path {
		t.Errorf("GetKirbi() = %q, want %q", got, path)
	}
	if !c.CanUseKirbi() {
		t.Error("CanUseKirbi() = false after a readable kirbi is set")
	}
}

func TestSetKirbiRejectsMissingFile(t *testing.T) {
	c := &credentials.Credentials{}
	if err := c.SetKirbi(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("SetKirbi() with a missing file error = nil, want an error")
	}
	if c.CanUseKirbi() {
		t.Error("CanUseKirbi() = true after a rejected path")
	}
}
