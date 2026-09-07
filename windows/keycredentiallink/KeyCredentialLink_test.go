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
	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/usage"
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

// msDS-KeyCredentialLink is read off objects the caller does not control, so a
// malformed value has to come back as an error from ParseDNWithBinary instead of
// ending the process. Each value below is well formed as a DN-with-binary and
// malformed as a blob.
func TestParseDNWithBinary_MalformedValues(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{
			// B:4:0000 declares four hexadecimal digits and supplies four, which is two
			// bytes: too few to hold the four-byte version.
			name:      "binary shorter than the version",
			value:     "B:4:0000:CN=PC01,DC=MANTICORE,DC=local",
			wantError: true,
		},
		{
			// A timestamp entry whose value cannot hold a 64-bit integer.
			name:      "truncated last logon timestamp",
			value:     "B:30:000200000100050104000800000000:CN=PC01,DC=MANTICORE,DC=local",
			wantError: true,
		},
		{
			// A version and nothing else: the entries the specification marks mandatory
			// are absent, which the caller handles, so this parses.
			name:      "version only",
			value:     "B:8:00020000:CN=PC01,DC=MANTICORE,DC=local",
			wantError: false,
		},
		{
			// A timestamp entry with no KeySource entry before it. The source governs how
			// the timestamp is decoded and is itself optional, so its absence must not be
			// dereferenced.
			name:      "last logon timestamp without a key source",
			value:     "B:30:000200000800080000000000000000:CN=PC01,DC=MANTICORE,DC=local",
			wantError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dnWithBinary := ldap.DNWithBinary{}
			if _, err := dnWithBinary.Unmarshal([]byte(test.value)); err != nil {
				t.Fatalf("DNWithBinary.Unmarshal() error = %v, want nil", err)
			}

			kc := keycredentiallink.KeyCredentialLink{}
			err := kc.ParseDNWithBinary(dnWithBinary)

			if test.wantError && err == nil {
				t.Fatalf("ParseDNWithBinary() returned no error, want one")
			}
			if !test.wantError && err != nil {
				t.Fatalf("ParseDNWithBinary() error = %v, want nil", err)
			}

			if err != nil {
				return
			}

			// A value that parses must also describe: the entries it lacks are reported,
			// not dereferenced.
			kc.Describe(0)
		})
	}
}

// A blob that omits the mandatory KeyMaterial entry cannot be marshalled back, and
// the integrity check that goes through the same path has to report a mismatch
// rather than dereferencing the absent material.
func TestToKeyCredentialLinkBlob_NoKeyMaterial(t *testing.T) {
	kc := keycredentiallink.KeyCredentialLink{
		Version: version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		KeyHash: []byte{0x01, 0x02, 0x03, 0x04},
	}

	if _, err := kc.ToKeyCredentialLinkBlob(); err == nil {
		t.Errorf("ToKeyCredentialLinkBlob() returned no error, want one")
	}

	if kc.CheckIntegrity() {
		t.Errorf("CheckIntegrity() = true, want false for a credential without key material")
	}
}

// Describe on a zero-valued structure touches every optional field at once.
func TestDescribe_ZeroValue(t *testing.T) {
	kc := keycredentiallink.KeyCredentialLink{}

	kc.Describe(0)
}

// KeyUsage_AdminKey is 0, which is also the zero value of the KeyUsage struct, so a
// blob carrying no KeyUsage entry used to be indistinguishable from one declaring an
// admin (PIN-reset) key — both when described and when marshalled back out.
func TestKeyUsage_AbsentEntryIsNotAdminKey(t *testing.T) {
	// A version plus a single KeyID entry: no KeyUsage entry at all.
	dnWithBinary := ldap.DNWithBinary{}
	if _, err := dnWithBinary.Unmarshal([]byte("B:18:00020000020001aabb:CN=PC01,DC=MANTICORE,DC=local")); err != nil {
		t.Fatalf("DNWithBinary.Unmarshal() error = %v, want nil", err)
	}

	kc := keycredentiallink.KeyCredentialLink{}
	if err := kc.ParseDNWithBinary(dnWithBinary); err != nil {
		t.Fatalf("ParseDNWithBinary() error = %v, want nil", err)
	}

	if kc.Usage != nil {
		t.Errorf("Usage = %v, want nil for a blob that carried no KeyUsage entry", kc.Usage)
	}

	// And it must describe without claiming a usage.
	kc.Describe(0)
}

// The absence must not be written back out as a declared usage. Marshalling needs
// key material, which the blob above does not carry, so this covers the same absent
// Usage on a credential that can be marshalled.
func TestKeyUsage_AbsentEntryIsNotMarshalled(t *testing.T) {
	kc := keycredentiallink.KeyCredentialLink{
		Version:     version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		KeyMaterial: testKeyMaterial(0x01),
	}

	if kc.Usage != nil {
		t.Fatalf("Usage = %v, want nil", kc.Usage)
	}

	blob, err := kc.ToKeyCredentialLinkBlob()
	if err != nil {
		t.Fatalf("ToKeyCredentialLinkBlob() error = %v, want nil", err)
	}

	for _, entry := range blob.Entries {
		if entry.Identifier == keycredentiallink.KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyUsage {
			t.Errorf("blob carries a KeyUsage entry with value %v, want none", entry.Value)
		}
	}
}

// A blob that does declare a usage keeps it, so the fix does not turn every usage
// into an absence.
func TestKeyUsage_DeclaredAdminKeyIsPreserved(t *testing.T) {
	// version 0x200 + KeyUsage entry (identifier 0x04) whose single byte is 0x00.
	dnWithBinary := ldap.DNWithBinary{}
	if _, err := dnWithBinary.Unmarshal([]byte("B:16:0002000001000400:CN=PC01,DC=MANTICORE,DC=local")); err != nil {
		t.Fatalf("DNWithBinary.Unmarshal() error = %v, want nil", err)
	}

	kc := keycredentiallink.KeyCredentialLink{}
	if err := kc.ParseDNWithBinary(dnWithBinary); err != nil {
		t.Fatalf("ParseDNWithBinary() error = %v, want nil", err)
	}

	if kc.Usage == nil {
		t.Fatalf("Usage = nil, want a declared AdminKey usage")
	}

	if kc.Usage.Value != usage.KeyUsage_AdminKey {
		t.Errorf("Usage.Value = 0x%02x, want 0x%02x (AdminKey)", kc.Usage.Value, usage.KeyUsage_AdminKey)
	}
}

// A legacy blob carries its usage as a string in the same entry identifier. Emitting
// the binary entry unconditionally gave such a blob two entries under identifier
// 0x04: a fabricated AdminKey plus the legacy string.
func TestKeyUsage_LegacyUsageDoesNotGainASecondEntry(t *testing.T) {
	kc := keycredentiallink.KeyCredentialLink{
		Version:     version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2},
		KeyMaterial: testKeyMaterial(0x01),
		LegacyUsage: "NGC",
	}

	blob, err := kc.ToKeyCredentialLinkBlob()
	if err != nil {
		t.Fatalf("ToKeyCredentialLinkBlob() error = %v, want nil", err)
	}

	usageEntries := make([][]byte, 0)
	for _, entry := range blob.Entries {
		if entry.Identifier == keycredentiallink.KEYCREDENTIALLINK_ENTRY_IDENTIFIER_KeyUsage {
			usageEntries = append(usageEntries, entry.Value)
		}
	}

	if len(usageEntries) != 1 {
		t.Errorf("blob carries %d KeyUsage entries (%v), want exactly 1 (the legacy string)", len(usageEntries), usageEntries)
	}

	if len(usageEntries) == 1 && string(usageEntries[0]) != "NGC" {
		t.Errorf("KeyUsage entry = %q, want %q", usageEntries[0], "NGC")
	}
}

// A credential built by the constructor still declares NGC.
func TestKeyUsage_ConstructorSetsNGC(t *testing.T) {
	now := utils.NewDateTimeFromTicks(0)
	kcv := version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}

	kc := keycredentiallink.NewKeyCredentialLink(kcv, "", testKeyMaterial(0x01), guid.NewGUID(), &now, &now)

	if kc.Usage == nil {
		t.Fatalf("Usage = nil, want NGC")
	}

	if kc.Usage.Value != usage.KeyUsage_NGC {
		t.Errorf("Usage.Value = 0x%02x, want 0x%02x (NGC)", kc.Usage.Value, usage.KeyUsage_NGC)
	}
}
