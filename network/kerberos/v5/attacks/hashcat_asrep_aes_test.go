package attacks

import (
	"encoding/hex"
	"strings"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// Example hashes of hashcat modes 32100 and 32200 (ST_HASH in
// src/modules/module_32100.c and module_32200.c), both for the password "hashcat",
// user "user" and realm "EXAMPLE.COM".
const (
	exASREP17 = "$krb5asrep$17$user$EXAMPLE.COM$a419c4030e555734b06c2629$c09a1421f96eb126c757a4b87830381f142477d9a85b2beb3093dbfd44f38ddb6016a479537fb7b36e046315869fe79187217971ff6a12c1e0a2df3f68045e03814b21f756d8981f781803d65e8572823c88979581d93cf7d768f2efced16f3719b8d1004d9e73d798de255383476bced47d1982f16be77d0feb55a1f44f58bd013fa4caee58ac614caf0f1cf9101ec9623c5b8c2a1491b73f134f074790088fdb360b5ebce0d32a8145ed00a81ddf77188e150b92d8e8ddd0285d27f1514253e5546e6bba864b362bb1e6483b26d08fa4cc268bfbefe0f690039bcc524b774599df3680c1c3431d891bfa99514a877f964e"
	exASREP18 = "$krb5asrep$18$user$EXAMPLE.COM$aa4c494f520b27873a4de8f7$ebc9976a77f62e8ccca02d43d68bafcc66a81fcbb44a336b00ce401982f32975a5f9bcdc752643252185866685b0a30aaf50e449e392a5994e6979f23aba25f7704c90b2efa03b703c3c2f9e3617cc588ed226d0417e7742d45407878fd946d046b4a9732b9a203cb857811714b009c195b7c96b9bccb7e48832b11a4e92ecf24c49e54de8d0d5d5351445b5126db90bb7eebc7861db1e61de1175824b0a45023a6fa06c2a9d3035fdcf863bea922648e3dc28b48e39b1dec0869e7fe4de399cb52dfcf2596599da54a4bb0169c72d9496de2e137a4594e0e8a69082fc558ac9ace65d32eae5e260a65ca3f2f5871aaeee7a3b090b50f39321d120c144421e0abe7d"
)

// asrepAESCipher rebuilds the AS-REP ciphertext (edata || checksum) from a hashcat
// "$krb5asrep$<etype>$user$realm$<checksum>$<edata>" hash.
func asrepAESCipher(t *testing.T, hash string) []byte {
	t.Helper()
	fields := strings.Split(hash, "$")
	if len(fields) != 7 {
		t.Fatalf("unexpected hashcat AS-REP layout: %d fields", len(fields))
	}
	cksum, err := hex.DecodeString(fields[5])
	if err != nil {
		t.Fatal(err)
	}
	edata, err := hex.DecodeString(fields[6])
	if err != nil {
		t.Fatal(err)
	}
	return append(edata, cksum...)
}

// TestFormatASREPHashAES checks etypes 17 and 18 reproduce hashcat's own example
// hashes for modes 32100 and 32200 byte for byte.
func TestFormatASREPHashAES(t *testing.T) {
	for _, tc := range []struct {
		etype int
		want  string
	}{
		{iana.ETypeAES128CTSHMACSHA196, exASREP17},
		{iana.ETypeAES256CTSHMACSHA196, exASREP18},
	} {
		got, err := FormatASREPHash("user", "EXAMPLE.COM", tc.etype, asrepAESCipher(t, tc.want))
		if err != nil {
			t.Fatalf("etype %d: %v", tc.etype, err)
		}
		if got != tc.want {
			t.Errorf("etype %d hash mismatch:\n got  %s\n want %s", tc.etype, got, tc.want)
		}
	}
}

// TestASREPAESExampleUsesDefaultSalt decrypts hashcat's example ciphertexts with
// the key derived from "hashcat" and the salt hashcat builds from the hash
// (upper-cased realm followed by the user name), which is what lets
// FormatASREPHash leave the salt out of the hash.
func TestASREPAESExampleUsesDefaultSalt(t *testing.T) {
	for _, tc := range []struct {
		etype int
		hash  string
	}{
		{iana.ETypeAES128CTSHMACSHA196, exASREP17},
		{iana.ETypeAES256CTSHMACSHA196, exASREP18},
	} {
		key, err := kerbcrypto.StringToKey(tc.etype, "hashcat", "EXAMPLE.COMuser", nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := kerbcrypto.Decrypt(tc.etype, key, kerbcrypto.KeyUsageASRepEncPart, asrepAESCipher(t, tc.hash)); err != nil {
			t.Errorf("etype %d: example does not decrypt under the default salt: %v", tc.etype, err)
		}
	}
}

// TestFormatASREPHashWithSalt checks a non-default salt is carried in the
// "$krb5asrep$<etype>$<salt>$<edata>$<checksum>" form modes 32100/32200 accept,
// and that the default salt, an empty salt and RC4 keep the plain form.
func TestFormatASREPHashWithSalt(t *testing.T) {
	cipher := asrepAESCipher(t, exASREP18)
	fields := strings.Split(exASREP18, "$")
	cksum, edata := fields[5], fields[6]

	got, err := FormatASREPHashWithSalt("user", "EXAMPLE.COM", "EXAMPLE.COMolduser", iana.ETypeAES256CTSHMACSHA196, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if want := "$krb5asrep$18$EXAMPLE.COMolduser$" + edata + "$" + cksum; got != want {
		t.Errorf("non-default salt:\n got  %s\n want %s", got, want)
	}

	for _, salt := range []string{"", "EXAMPLE.COMuser"} {
		got, err := FormatASREPHashWithSalt("user", "example.com", salt, iana.ETypeAES256CTSHMACSHA196, cipher)
		if err != nil {
			t.Fatal(err)
		}
		if want := "$krb5asrep$18$user$example.com$" + cksum + "$" + edata; got != want {
			t.Errorf("salt %q:\n got  %s\n want %s", salt, got, want)
		}
	}

	rc4, err := FormatASREPHashWithSalt("user", "domain.com", "ignored", iana.ETypeRC4HMAC, make([]byte, 64))
	if err != nil || !strings.HasPrefix(rc4, "$krb5asrep$23$user@domain.com:") {
		t.Errorf("RC4 with a salt: %q, %v", rc4, err)
	}
}

// TestFormatASREPHashRejectsUnsupportedAES checks the SHA-2 AES enctypes, which
// hashcat has no AS-REP mode for, and ciphertexts too short to be an AS-REP are
// refused.
func TestFormatASREPHashRejectsUnsupportedAES(t *testing.T) {
	for _, etype := range []int{iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384, 99} {
		if _, err := FormatASREPHash("user", "domain.com", etype, make([]byte, 128)); err == nil {
			t.Errorf("etype %d accepted", etype)
		}
	}
	if _, err := FormatASREPHash("user", "domain.com", iana.ETypeAES256CTSHMACSHA196, make([]byte, 12+31)); err == nil {
		t.Error("31 bytes of AES encrypted data accepted")
	}
}

// TestFormatASREPHashRefusesDollarFields checks a '$' in the fields hashcat splits
// on is refused rather than producing a hash hashcat rejects with a token-length
// error, and that a computer account (name ending in '$') still formats through
// its KDC-reported salt, which has none.
func TestFormatASREPHashRefusesDollarFields(t *testing.T) {
	cipher := asrepAESCipher(t, exASREP18)
	if _, err := FormatASREPHash("svc$", "EXAMPLE.COM", iana.ETypeAES256CTSHMACSHA196, cipher); err == nil {
		t.Error("FormatASREPHash accepted a user name containing '$'")
	}
	if _, err := FormatASREPHashWithSalt("svc$", "EXAMPLE.COM", "EXAMPLE.COMsvc$", iana.ETypeAES256CTSHMACSHA196, cipher); err == nil {
		t.Error("FormatASREPHashWithSalt accepted a default salt containing '$'")
	}
	if _, err := FormatASREPHashWithSalt("svc", "EXAMPLE.COM", "EXAMPLE.COMold$svc", iana.ETypeAES256CTSHMACSHA196, cipher); err == nil {
		t.Error("FormatASREPHashWithSalt accepted a salt containing '$'")
	}
	got, err := FormatASREPHashWithSalt("HOST$", "EXAMPLE.COM", "EXAMPLE.COMhosthost.example.com", iana.ETypeAES256CTSHMACSHA196, cipher)
	if err != nil {
		t.Fatalf("computer account with its KDC salt: %v", err)
	}
	if !strings.HasPrefix(got, "$krb5asrep$18$EXAMPLE.COMhosthost.example.com$") {
		t.Errorf("computer account hash %.60s..., want the salt-carrying form", got)
	}
}
