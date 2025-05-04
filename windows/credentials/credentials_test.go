package credentials_test

import (
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
