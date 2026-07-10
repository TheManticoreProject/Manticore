package attacks

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// John the Ripper's krb5asrep/krb5tgs formats differ from hashcat's in a few
// subtle-but-load-bearing ways (verified against John Jumbo's own tests[]
// vectors in src/krb5_asrep_fmt_plug.c and src/krb5_tgs_common_plug.c):
//
//   - RC4 (etype 23) AS-REP: John drops the "user@realm:" salt that hashcat mode
//     18200 carries and stores only "$krb5asrep$23$<checksum>$<edata>" (RC4 has
//     no salt, so the principal is not needed to crack).
//   - RC4 (etype 23) TGS: John matches hashcat mode 13100 exactly —
//     "$krb5tgs$23$*<user>$<realm>$<spn>*$<checksum>$<edata>".
//   - AES (etype 17/18): both AS-REP and TGS use "$krb5<kind>$<etype>$<salt>$
//     <edata>$<checksum>", where the checksum (the truncated HMAC) is appended
//     AFTER the encrypted data (the reverse of RC4, where it comes first) and the
//     salt is the Kerberos string-to-key salt "<UPPER-REALM><account>" with no
//     separator (e.g. "EXAMPLE.COMluser").
//
// These functions sit alongside the hashcat formatters (FormatASREPHash /
// FormatTGSHash) and reuse the same splitCipher helper to separate the checksum
// from the encrypted data.

// aesSalt builds the John/Kerberos string-to-key salt for an AES roast hash:
// the uppercased realm concatenated with the account name and no separator, as
// John expects it (e.g. realm "example.com" + user "luser" -> "EXAMPLE.COMluser").
// This is correct for user service accounts (the common Kerberoasting /
// AS-REP-roasting target); machine accounts use a host-based salt that the caller
// must supply explicitly if it differs.
func aesSalt(account, realm string) string {
	return strings.ToUpper(realm) + account
}

// FormatASREPHashJohn formats an AS-REP encrypted part as a John the Ripper
// krb5asrep hash.
//
// RC4 (etype 23) yields "$krb5asrep$23$<checksum>$<edata>" (no principal, unlike
// hashcat mode 18200). AES enctypes (17/18) yield
// "$krb5asrep$<etype>$<UPPER-REALM><username>$<edata>$<checksum>".
func FormatASREPHashJohn(username, realm string, etype int, cipher []byte) (string, error) {
	cksum, edata, err := splitCipher(etype, cipher)
	if err != nil {
		return "", err
	}
	if etype == iana.ETypeRC4HMAC {
		return fmt.Sprintf("$krb5asrep$23$%s$%s",
			hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
	}
	return fmt.Sprintf("$krb5asrep$%d$%s$%s$%s",
		etype, aesSalt(username, realm), hex.EncodeToString(edata), hex.EncodeToString(cksum)), nil
}

// FormatTGSHashJohn formats a service ticket's encrypted part as a John the
// Ripper krb5tgs hash.
//
// RC4 (etype 23) yields "$krb5tgs$23$*<account>$<realm>$<spn>*$<checksum>$<edata>"
// (identical to hashcat mode 13100). AES enctypes (17/18) yield
// "$krb5tgs$<etype>$<UPPER-REALM><account>$<edata>$<checksum>" (the checksum is
// appended after the encrypted data and the account is folded into the salt).
func FormatTGSHashJohn(account, realm, spn string, etype int, cipher []byte) (string, error) {
	cksum, edata, err := splitCipher(etype, cipher)
	if err != nil {
		return "", err
	}
	if etype == iana.ETypeRC4HMAC {
		return fmt.Sprintf("$krb5tgs$23$*%s$%s$%s*$%s$%s",
			account, realm, spn, hex.EncodeToString(cksum), hex.EncodeToString(edata)), nil
	}
	return fmt.Sprintf("$krb5tgs$%d$%s$%s$%s",
		etype, aesSalt(account, realm), hex.EncodeToString(edata), hex.EncodeToString(cksum)), nil
}
