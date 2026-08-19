package kerberos

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/keytab"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credentials"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// Keytab authentication (pass-the-key from a .keytab)
//
// A keytab stores an account's long-term keys, so it lets a client authenticate
// non-interactively — no password. These methods select a key from a keytab and
// wire it into the client as an AES pass-the-key (etype 17-20) or, for RC4, an
// overpass-the-hash NT-hash credential, exactly as WithAESKey / WithNTHash would.
// GetTGT / GetTGS then run unchanged.

// credentialFromKeytabEntry converts a selected keytab entry into a long-term
// credential for username@realm. AES128/AES256 keys become pass-the-key
// credentials; the RC4-HMAC key is the NT hash, so it becomes an
// overpass-the-hash credential. Other enctypes such as DES are not supported.
func credentialFromKeytabEntry(e *keytab.Entry, username, realm string) (*credentials.Credential, error) {
	switch int(e.EType) {
	case iana.ETypeAES128CTSHMACSHA196, iana.ETypeAES256CTSHMACSHA196,
		iana.ETypeAES128CTSHMACSHA256, iana.ETypeAES256CTSHMACSHA384:
		return credentials.NewWithAESKey(username, realm, int(e.EType), e.Key)
	case iana.ETypeRC4HMAC:
		return credentials.NewWithNTHash(username, realm, e.Key)
	default:
		return nil, fmt.Errorf("kerberos: keytab enctype %d is not usable for authentication (need AES or RC4)", e.EType)
	}
}

// WithKeytab selects a long-term key from a parsed keytab and configures it as
// the client's credential (pass-the-key), so GetTGT needs no password.
//
// The entry is chosen for the client's own principal (username@realm); the
// strongest usable enctype present is preferred (AES256 > AES128 > RC4). When the
// client was created with an empty username (NewClient("", realm, kdc)), the
// principal of the selected entry populates the client's username. Pass etype > 0
// to force a specific enctype instead of the strongest.
func (c *KerberosClient) WithKeytab(kt *keytab.Keytab, etype int) error {
	principal := ""
	if c.username != "" {
		principal = c.username + "@" + c.realm
	}
	entry := kt.Select(principal, etype, -1)
	if entry == nil {
		return fmt.Errorf("kerberos: no keytab entry for %q (etype %d)", principal, etype)
	}

	username := c.username
	realm := c.realm
	if username == "" {
		username = entry.Principal.String()
		if len(entry.Principal.Components) > 0 {
			username = entry.Principal.Components[0]
		}
		if realm == "" {
			realm = entry.Principal.Realm
		}
	}

	cred, err := credentialFromKeytabEntry(entry, username, realm)
	if err != nil {
		return err
	}
	c.username = username
	c.realm = cred.Realm()
	c.cred = cred
	return nil
}

// WithKeytabBytes parses keytab bytes and configures the client's credential
// from them via WithKeytab (strongest usable enctype, etype 0).
func (c *KerberosClient) WithKeytabBytes(data []byte) error {
	kt, err := keytab.Unmarshal(data)
	if err != nil {
		return err
	}
	return c.WithKeytab(kt, 0)
}

// WithKeytabFile reads a .keytab file and configures the client's credential
// from it (strongest usable enctype).
func (c *KerberosClient) WithKeytabFile(path string) error {
	kt, err := keytab.Load(path)
	if err != nil {
		return err
	}
	return c.WithKeytab(kt, 0)
}
