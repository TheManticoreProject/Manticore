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
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/utils"
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

// testKeyMaterial builds a BCRYPT_RSA_PUBLIC_KEY with a deterministic modulus, so the
// derived KeyID is stable across runs.
func testKeyMaterial(modulusByte byte) *keys.BCRYPT_RSA_PUBLIC_KEY {
	exponent := []byte{0x01, 0x00, 0x01}
	modulus := make([]byte, 256)
	for i := range modulus {
		modulus[i] = modulusByte
	}

	return &keys.BCRYPT_RSA_PUBLIC_KEY{
		Magic: magic.BCRYPT_KEY_BLOB{Magic: magic.BCRYPT_RSAPUBLIC_MAGIC},
		Header: headers.BCRYPT_RSA_KEY_BLOB{
			BitLength:   uint32(len(modulus) * 8),
			CbPublicExp: uint32(len(exponent)),
			CbModulus:   uint32(len(modulus)),
		},
		Content: blob.BCRYPT_RSA_PUBLIC_BLOB{
			PublicExponent: exponent,
			Modulus:        modulus,
		},
	}
}

// An empty identifier has to be filled in with the KeyID the key material requires.
// The KDC looks a PKINIT key up by that hash, so a credential built without one would
// otherwise carry an empty KeyID and never match the account.
func TestNewKeyCredentialLink_DerivesIdentifierWhenEmpty(t *testing.T) {
	keyMaterial := testKeyMaterial(0xAB)
	now := utils.NewDateTimeFromTicks(0)

	for _, kcv := range []uint32{
		version.KeyCredentialLinkVersion_0,
		version.KeyCredentialLinkVersion_1,
		version.KeyCredentialLinkVersion_2,
	} {
		kc := keycredentiallink.NewKeyCredentialLink(
			version.KeyCredentialLinkVersion{Value: kcv},
			"",
			keyMaterial,
			guid.NewGUID(),
			&now,
			&now,
		)

		rawKeyMaterial, err := keyMaterial.Marshal()
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		want := utils.ComputeKeyIdentifier(rawKeyMaterial, version.KeyCredentialLinkVersion{Value: kcv})

		if kc.Identifier == "" {
			t.Errorf("version %v: Identifier is empty, want the derived KeyID", kcv)
		}
		if kc.Identifier != want {
			t.Errorf("version %v: Identifier = %s, want %s", kcv, kc.Identifier, want)
		}
	}
}

// A caller that supplies an identifier is reproducing a specific credential, so the
// value has to survive untouched.
func TestNewKeyCredentialLink_KeepsExplicitIdentifier(t *testing.T) {
	const explicit = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="
	now := utils.NewDateTimeFromTicks(0)

	kc := keycredentiallink.NewKeyCredentialLink(
		version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		explicit,
		testKeyMaterial(0xCD),
		guid.NewGUID(),
		&now,
		&now,
	)

	if kc.Identifier != explicit {
		t.Errorf("Identifier = %s, want the supplied %s", kc.Identifier, explicit)
	}
}

// Different key material has to yield a different KeyID, otherwise the KDC could not
// tell two credentials apart.
func TestComputeKeyIdentifier_DistinguishesKeys(t *testing.T) {
	now := utils.NewDateTimeFromTicks(0)
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}

	first := keycredentiallink.NewKeyCredentialLink(kcv, "", testKeyMaterial(0x01), guid.NewGUID(), &now, &now)
	second := keycredentiallink.NewKeyCredentialLink(kcv, "", testKeyMaterial(0x02), guid.NewGUID(), &now, &now)

	if first.Identifier == second.Identifier {
		t.Errorf("two different keys produced the same KeyID: %s", first.Identifier)
	}
}

// Key material that cannot be marshalled, or is absent, must not panic.
func TestComputeKeyIdentifier_NoKeyMaterial(t *testing.T) {
	kc := keycredentiallink.KeyCredentialLink{
		Version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
	}

	if got := kc.ComputeKeyIdentifier(); got != "" {
		t.Errorf("ComputeKeyIdentifier() = %q, want an empty string", got)
	}
}
