package attacks

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// TestFormatASREPHashJohnRC4 checks the RC4 (etype 23) AS-REP form against a
// John Jumbo tests[] vector: "$krb5asrep$23$<checksum>$<edata>" with no
// principal (RC4 has no salt), and the checksum taken from the FIRST 16 bytes.
func TestFormatASREPHashJohnRC4(t *testing.T) {
	cksum := "c447eddaebf22ebf006a8fc6f986488c"
	edata := "eb3a17eb56287b474cecad5d4e0490d9"
	cipher := append(mustHex(t, cksum), mustHex(t, edata)...)

	got, err := FormatASREPHashJohn("victim", "corp.local", iana.ETypeRC4HMAC, cipher)
	if err != nil {
		t.Fatalf("FormatASREPHashJohn: %v", err)
	}
	want := "$krb5asrep$23$" + cksum + "$" + edata
	if got != want {
		t.Errorf("RC4 AS-REP John:\n got  %s\n want %s", got, want)
	}
}

// TestFormatASREPHashJohnAES checks the AES (etype 18) AS-REP form against the
// John Jumbo tests[] vector "$krb5asrep$18$EXAMPLE.COMluser$<edata>$<checksum>":
// salt = UPPER(realm)+user, edata BEFORE the trailing 12-byte HMAC checksum.
func TestFormatASREPHashJohnAES(t *testing.T) {
	edata := "42e34732112be6cec1532177a6c93af5"
	cksum := "420973360c2e907b9053f1db" // 12 bytes (aes256-cts-hmac-sha1-96)
	cipher := append(mustHex(t, edata), mustHex(t, cksum)...)

	got, err := FormatASREPHashJohn("luser", "example.com", iana.ETypeAES256CTSHMACSHA196, cipher)
	if err != nil {
		t.Fatalf("FormatASREPHashJohn: %v", err)
	}
	want := "$krb5asrep$18$EXAMPLE.COMluser$" + edata + "$" + cksum
	if got != want {
		t.Errorf("AES AS-REP John:\n got  %s\n want %s", got, want)
	}
}

// TestFormatTGSHashJohnRC4 checks the RC4 (etype 23) TGS form, which John shares
// with hashcat mode 13100: "$krb5tgs$23$*user$realm$spn*$<checksum>$<edata>".
func TestFormatTGSHashJohnRC4(t *testing.T) {
	cksum := "63386d22d359fe42230300d56852c9eb"
	edata := "0011223344556677"
	cipher := append(mustHex(t, cksum), mustHex(t, edata)...)

	got, err := FormatTGSHashJohn("user", "realm", "test/spn", iana.ETypeRC4HMAC, cipher)
	if err != nil {
		t.Fatalf("FormatTGSHashJohn: %v", err)
	}
	want := "$krb5tgs$23$*user$realm$test/spn*$" + cksum + "$" + edata
	if got != want {
		t.Errorf("RC4 TGS John:\n got  %s\n want %s", got, want)
	}
}

// TestFormatTGSHashJohnAES checks the AES (etype 18) TGS form:
// "$krb5tgs$18$<UPPER-REALM><account>$<edata>$<checksum>" with the checksum
// appended after the encrypted data (mirroring the AES AS-REP layout).
func TestFormatTGSHashJohnAES(t *testing.T) {
	edata := "8899aabbccddeeff0011223344556677"
	cksum := "aabbccddeeff00112233445566"[:24] // 12 bytes
	cipher := append(mustHex(t, edata), mustHex(t, cksum)...)

	got, err := FormatTGSHashJohn("svc_web", "corp.local", "HTTP/web.corp.local", iana.ETypeAES256CTSHMACSHA196, cipher)
	if err != nil {
		t.Fatalf("FormatTGSHashJohn: %v", err)
	}
	want := "$krb5tgs$18$CORP.LOCALsvc_web$" + edata + "$" + cksum
	if got != want {
		t.Errorf("AES TGS John:\n got  %s\n want %s", got, want)
	}
}

// TestFormatJohnUnsupportedEType ensures an unsupported enctype is rejected.
func TestFormatJohnUnsupportedEType(t *testing.T) {
	if _, err := FormatASREPHashJohn("u", "r", 1, []byte{1, 2, 3}); err == nil {
		t.Error("expected error for unsupported etype in AS-REP John format")
	}
	if _, err := FormatTGSHashJohn("u", "r", "s/h", 1, []byte{1, 2, 3}); err == nil {
		t.Error("expected error for unsupported etype in TGS John format")
	}
}
