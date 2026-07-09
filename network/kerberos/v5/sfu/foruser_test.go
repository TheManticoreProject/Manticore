package sfu

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func impersonated() messages.PrincipalName {
	return messages.PrincipalName{NameType: iana.NameTypePrincipal, NameString: []string{"administrator"}}
}

func TestBuildParsePAForUserRoundtrip(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32) // AES256 TGT session key
	pa, err := BuildPAForUser(impersonated(), "CORP.LOCAL", key, iana.ETypeAES256CTSHMACSHA196)
	if err != nil {
		t.Fatalf("BuildPAForUser: %v", err)
	}
	if pa.PADataType != iana.PAForUser {
		t.Errorf("padata type = %d, want %d", pa.PADataType, iana.PAForUser)
	}

	p, err := ParsePAForUser(pa.PADataValue)
	if err != nil {
		t.Fatalf("ParsePAForUser: %v", err)
	}
	if len(p.UserName.NameString) != 1 || p.UserName.NameString[0] != "administrator" {
		t.Errorf("userName mismatch: %+v", p.UserName)
	}
	if p.UserRealm != "CORP.LOCAL" {
		t.Errorf("userRealm = %q", p.UserRealm)
	}
	if p.AuthPackage != "Kerberos" {
		t.Errorf("auth-package = %q, want Kerberos", p.AuthPackage)
	}
	if p.Cksum.CKSumType != iana.CksumTypeHMACSHA196AES256 {
		t.Errorf("cksum type = %d, want %d (paired with AES256)", p.Cksum.CKSumType, iana.CksumTypeHMACSHA196AES256)
	}
	if !VerifyPAForUser(p, key, iana.ETypeAES256CTSHMACSHA196) {
		t.Error("VerifyPAForUser rejected a valid element")
	}
}

func TestPAForUserChecksumTypeFollowsSessionKey(t *testing.T) {
	// RC4 session key -> KERB_CHECKSUM_HMAC_MD5 (the [MS-SFU] literal type).
	rc4Key := bytes.Repeat([]byte{0x11}, 16)
	pa, err := BuildPAForUser(impersonated(), "CORP.LOCAL", rc4Key, iana.ETypeRC4HMAC)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := ParsePAForUser(pa.PADataValue)
	if p.Cksum.CKSumType != iana.CksumTypeHMACMD5 {
		t.Errorf("RC4 session key should use HMAC-MD5 checksum (-138), got %d", p.Cksum.CKSumType)
	}
	if !VerifyPAForUser(p, rc4Key, iana.ETypeRC4HMAC) {
		t.Error("RC4 PA-FOR-USER did not verify")
	}
}

func TestVerifyPAForUserDetectsTamper(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	pa, _ := BuildPAForUser(impersonated(), "CORP.LOCAL", key, iana.ETypeAES256CTSHMACSHA196)
	p, _ := ParsePAForUser(pa.PADataValue)

	// Change the impersonated user; the checksum must no longer match.
	p.UserName.NameString[0] = "guest"
	if VerifyPAForUser(p, key, iana.ETypeAES256CTSHMACSHA196) {
		t.Error("VerifyPAForUser accepted a tampered impersonation target")
	}

	// Wrong session key must also fail.
	p2, _ := ParsePAForUser(pa.PADataValue)
	if VerifyPAForUser(p2, bytes.Repeat([]byte{0x99}, 32), iana.ETypeAES256CTSHMACSHA196) {
		t.Error("VerifyPAForUser accepted the wrong session key")
	}
}

func TestBuildPAForUserRejectsBadEType(t *testing.T) {
	if _, err := BuildPAForUser(impersonated(), "R", bytes.Repeat([]byte{1}, 16), 999); err == nil {
		t.Error("expected error for unsupported session-key etype")
	}
}
