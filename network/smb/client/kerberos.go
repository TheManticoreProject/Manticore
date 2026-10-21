package client

import (
	"fmt"
	"os"
	"strings"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/keytab"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// KerberosOptions configures LoginKerberos.
type KerberosOptions struct {
	// KDCHost is the address of the KDC (a domain controller). When empty, the
	// KDC is discovered through the DNS SRV records of the realm.
	KDCHost string
	// SPN is the service principal name of the server. When empty it defaults
	// to "cifs/<host>", where host is the name Dial was given; dial by DNS name,
	// not by IP address, for the default to name a registered principal.
	SPN string
}

// LoginKerberos authenticates a session with Kerberos over SPNEGO. The realm is
// creds' domain, upper-cased, and the client principal its username. The first secret creds
// holds, in this order, obtains the credentials:
//
//   - a ccache (creds.CCache) or a .kirbi (creds.Kirbi): pass-the-ticket, the
//     realm and username are taken from the ticket;
//   - a keytab (creds.Keytab) or an AES key (creds.AESKey): pass-the-key;
//   - an NT hash: overpass-the-hash (RC4);
//   - a password.
//
// When creds holds none of them, the TGT is loaded from the ccache named by the
// KRB5CCNAME environment variable.
//
// Kerberos is supported over SMB 2.0.2 and later; over SMB1 LoginKerberos
// returns an error.
func (c *Client) LoginKerberos(creds *credentials.Credentials, opts KerberosOptions) error {
	if creds == nil {
		return fmt.Errorf("kerberos login requires credentials but none were provided")
	}
	kc, err := newKerberosClient(creds, opts.KDCHost)
	if err != nil {
		return err
	}
	spn := opts.SPN
	if spn == "" {
		spn = "cifs/" + c.host
	}
	return c.backend.LoginKerberos(kc, spn)
}

// newKerberosClient builds the native Kerberos client LoginKerberos
// authenticates with, from the first secret creds holds.
func newKerberosClient(creds *credentials.Credentials, kdcHost string) (*kerberos.KerberosClient, error) {
	// The domain of SMB credentials is a Windows domain name, which is
	// case-insensitive; its Kerberos realm is the upper-case DNS domain name
	// (as impacket and the Windows client form it), and the KDC replies with
	// that spelling.
	kc := kerberos.NewClient(creds.GetUsername(), strings.ToUpper(creds.GetDomain()), kdcHost)

	switch {
	case creds.GetCCache() != "":
		if err := kc.LoadTGTFromCCacheFile(creds.GetCCache()); err != nil {
			return nil, fmt.Errorf("kerberos: cannot load the TGT from ccache %q: %w", creds.GetCCache(), err)
		}
	case creds.GetKirbi() != "":
		if err := kc.LoadTGTFromKirbiFile(creds.GetKirbi()); err != nil {
			return nil, fmt.Errorf("kerberos: cannot load the TGT from %q: %w", creds.GetKirbi(), err)
		}
	case creds.GetKeytab() != "":
		kt, err := keytab.Load(creds.GetKeytab())
		if err != nil {
			return nil, fmt.Errorf("kerberos: cannot read keytab %q: %w", creds.GetKeytab(), err)
		}
		if err := kc.WithKeytab(kt, 0); err != nil {
			return nil, err
		}
	case creds.GetAESKey() != "":
		if err := kc.WithAESKey(creds.GetAESKey()); err != nil {
			return nil, fmt.Errorf("kerberos: invalid AES key: %w", err)
		}
	case creds.GetNTHash() != "":
		if err := kc.WithNTHash(creds.GetNTHash()); err != nil {
			return nil, fmt.Errorf("kerberos: invalid NT hash: %w", err)
		}
	case creds.GetPassword() != "":
		kc.WithPassword(creds.GetPassword())
	default:
		if os.Getenv("KRB5CCNAME") == "" {
			return nil, fmt.Errorf("kerberos: no secret to authenticate with: provide a password, NT hash, AES key, keytab, ccache or kirbi, or set KRB5CCNAME")
		}
		if err := kc.LoadTGTFromCCacheEnv(); err != nil {
			return nil, fmt.Errorf("kerberos: cannot load the TGT from KRB5CCNAME: %w", err)
		}
	}
	return kc, nil
}
