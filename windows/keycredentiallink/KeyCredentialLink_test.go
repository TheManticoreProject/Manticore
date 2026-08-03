package keycredentiallink_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/blob"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/magic"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink"
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

func TestKeyCredential_Unmarshal(t *testing.T) {
	tests := []struct {
		name                       string
		msDsKeyCredentialLinkValue string
		wantErr                    bool
	}{
		{
			name:                       "Valid KeyCredential with specific identifier",
			msDsKeyCredentialLinkValue: "B:10:9012345678:CN=POC,CN=Computers,DC=MANTICORE,DC=local",
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dnb := ldap.DNWithBinary{}
			bytesRead, err := dnb.Unmarshal([]byte(tt.msDsKeyCredentialLinkValue))
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unmarshal() error = %v", err)
				}
				return
			}
			if bytesRead != len([]byte(tt.msDsKeyCredentialLinkValue)) {
				t.Errorf("Unmarshal() bytesRead = %v, want %v", bytesRead, len([]byte(tt.msDsKeyCredentialLinkValue)))
				return
			}

			kcl := keycredentiallink.KeyCredentialLink{}
			_, err = kcl.Unmarshal(dnb.BinaryData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestKeyCredential_Unmarshal_PreservesVersion checks that Unmarshal carries the
// blob version onto the KeyCredentialLink. The version was previously left at its
// zero value, so a v2 credential round-tripped as v0 and the version-dependent
// entry parsing ran with the wrong version.
func TestKeyCredential_Unmarshal_PreservesVersion(t *testing.T) {
	built := keycredentiallink.NewKeyCredentialLink(
		version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		&keys.BCRYPT_RSA_PUBLIC_KEY{
			Magic:   magic.BCRYPT_KEY_BLOB{Magic: magic.BCRYPT_RSAPUBLIC_MAGIC},
			Header:  headers.BCRYPT_RSA_KEY_BLOB{BitLength: 16, CbPublicExp: 3, CbModulus: 2},
			Content: blob.BCRYPT_RSA_PUBLIC_BLOB{PublicExponent: []byte{0x01, 0x00, 0x01}, Modulus: []byte{0x00, 0x01}},
		},
		guid.NewGUID(),
		nil,
		nil,
	)

	raw, err := built.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	parsed := keycredentiallink.KeyCredentialLink{}
	if _, err := parsed.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if parsed.Version.Value != version.KeyCredentialLinkVersion_2 {
		t.Errorf("Version.Value = 0x%x, want 0x%x", parsed.Version.Value, version.KeyCredentialLinkVersion_2)
	}
}
