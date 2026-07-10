package pac

import (
	"encoding/binary"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
)

// PAC_CREDENTIAL_INFO / PAC_CREDENTIAL_DATA / NTLM_SUPPLEMENTAL_CREDENTIAL
// support ([MS-PAC] §2.6). This buffer is only present in a PAC produced from a
// PKINIT (certificate) logon and carries the account's NTLM secrets encrypted
// under the AS reply key — the basis of the "UnPAC-the-hash" technique.

// CredentialInfo is a parsed PAC_CREDENTIAL_INFO buffer ([MS-PAC] §2.6.1).
type CredentialInfo struct {
	// Version is the structure version (0).
	Version uint32
	// EncryptionType is the etype of SerializedData (the AS reply key's etype).
	EncryptionType uint32
	// SerializedData is the encrypted PAC_CREDENTIAL_DATA.
	SerializedData []byte
}

// ParseCredentialInfo parses a PAC_CREDENTIAL_INFO buffer.
func ParseCredentialInfo(data []byte) (*CredentialInfo, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("pac: PAC_CREDENTIAL_INFO too short (%d bytes)", len(data))
	}
	return &CredentialInfo{
		Version:        binary.LittleEndian.Uint32(data[0:]),
		EncryptionType: binary.LittleEndian.Uint32(data[4:]),
		SerializedData: append([]byte(nil), data[8:]...),
	}, nil
}

// NTLMCredential holds the LM and NT hashes recovered from a
// NTLM_SUPPLEMENTAL_CREDENTIAL ([MS-PAC] §2.6.4).
type NTLMCredential struct {
	// Version is the structure version (0).
	Version uint32
	// Flags indicates which hashes are present (bit 0 = LM, bit 1 = NT).
	Flags uint32
	// LMHash is the 16-byte LM hash (valid when Flags&1 != 0).
	LMHash []byte
	// NTHash is the 16-byte NT hash (valid when Flags&2 != 0).
	NTHash []byte
}

const (
	ntlmCredentialLMPresent = 0x00000001
	ntlmCredentialNTPresent = 0x00000002
)

// DecryptCredentialData decrypts a PAC_CREDENTIAL_INFO's SerializedData with the
// AS reply key (the PKINIT-derived key), yielding the NDR-serialized
// PAC_CREDENTIAL_DATA. The key usage is KERB_NON_KERB_SALT (16) per [MS-PAC].
func (ci *CredentialInfo) DecryptCredentialData(replyKey []byte) ([]byte, error) {
	plain, err := kerbcrypto.Decrypt(int(ci.EncryptionType), replyKey, kerbcrypto.KeyUsageKerbNonKerbSalt, ci.SerializedData)
	if err != nil {
		return nil, fmt.Errorf("pac: decrypt PAC_CREDENTIAL_DATA: %w", err)
	}
	return plain, nil
}

// ParseNTLMSupplementalCredential extracts the NTLM_SUPPLEMENTAL_CREDENTIAL from
// a decrypted PAC_CREDENTIAL_DATA blob.
//
// PAC_CREDENTIAL_DATA is NDR type-serialized ([MS-PAC] §2.6.2): a 16-byte type
// serialization header, then CredentialCount and an array of
// SECPKG_SUPPLEMENTAL_CRED. For the single "NTLM" package Windows emits, the
// deferred Credentials byte array is the 40-byte NTLM_SUPPLEMENTAL_CREDENTIAL
// (Version, Flags, LM[16], NT[16]); it is the final conformant array in the
// buffer, located here via its 4-byte MaximumCount prefix.
func ParseNTLMSupplementalCredential(data []byte) (*NTLMCredential, error) {
	const ntlmStructLen = 40 // 4 + 4 + 16 + 16
	// The NTLM_SUPPLEMENTAL_CREDENTIAL is the trailing conformant byte array,
	// prefixed by its 4-byte NDR MaximumCount. Scan from the end for a
	// MaximumCount of 40 whose 40 following bytes fit exactly in the buffer.
	for off := len(data) - ntlmStructLen - 4; off >= 0; off-- {
		if binary.LittleEndian.Uint32(data[off:]) != ntlmStructLen {
			continue
		}
		body := data[off+4:]
		if len(body) < ntlmStructLen {
			continue
		}
		cred := &NTLMCredential{
			Version: binary.LittleEndian.Uint32(body[0:]),
			Flags:   binary.LittleEndian.Uint32(body[4:]),
			LMHash:  append([]byte(nil), body[8:24]...),
			NTHash:  append([]byte(nil), body[24:40]...),
		}
		// A valid NTLM_SUPPLEMENTAL_CREDENTIAL has version 0 and at least one
		// hash flagged present.
		if cred.Version == 0 && cred.Flags&(ntlmCredentialLMPresent|ntlmCredentialNTPresent) != 0 {
			if cred.Flags&ntlmCredentialLMPresent == 0 {
				cred.LMHash = nil
			}
			if cred.Flags&ntlmCredentialNTPresent == 0 {
				cred.NTHash = nil
			}
			return cred, nil
		}
	}
	return nil, fmt.Errorf("pac: no NTLM_SUPPLEMENTAL_CREDENTIAL found in PAC_CREDENTIAL_DATA")
}
