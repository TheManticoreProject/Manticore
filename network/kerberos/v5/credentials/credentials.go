// Package credentials models the long-term secret a Kerberos client
// authenticates with, abstracting over the three forms Active Directory
// tooling uses: a cleartext password, an NT hash (the RC4-HMAC key — enabling
// overpass-the-hash), or a raw AES key (enabling pass-the-key). A Credential
// turns any of these into the per-etype key material the AS/TGS exchanges need.
package credentials

import (
	"encoding/hex"
	"fmt"
	"strings"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// SecretKind identifies which form of long-term secret a Credential holds.
type SecretKind int

const (
	// SecretPassword is a cleartext password; keys of any supported etype can
	// be derived from it via string-to-key.
	SecretPassword SecretKind = iota
	// SecretNTHash is the NT hash (MD4 of the UTF-16LE password), which is
	// exactly the RC4-HMAC (etype 23) key — the basis of overpass-the-hash.
	SecretNTHash
	// SecretAESKey is a precomputed AES key for one specific etype (17 or 18),
	// the basis of pass-the-key.
	SecretAESKey
)

// Credential is an immutable long-term secret bound to a principal.
type Credential struct {
	username string
	realm    string // always upper-cased

	kind     SecretKind
	password string
	ntHash   []byte // 16 bytes, when kind == SecretNTHash
	aesKey   []byte // 16 or 32 bytes, when kind == SecretAESKey
	aesEtype int    // 17 or 18, when kind == SecretAESKey
}

// NewWithPassword creates a password-based credential. The realm is upper-cased.
func NewWithPassword(username, realm, password string) *Credential {
	return &Credential{
		username: username,
		realm:    strings.ToUpper(realm),
		kind:     SecretPassword,
		password: password,
	}
}

// NewWithNTHash creates an NT-hash (overpass-the-hash) credential from a 16-byte
// hash. The hash is copied.
func NewWithNTHash(username, realm string, ntHash []byte) (*Credential, error) {
	if len(ntHash) != 16 {
		return nil, fmt.Errorf("credentials: NT hash must be 16 bytes, got %d", len(ntHash))
	}
	h := make([]byte, 16)
	copy(h, ntHash)
	return &Credential{
		username: username,
		realm:    strings.ToUpper(realm),
		kind:     SecretNTHash,
		ntHash:   h,
	}, nil
}

// NewWithHexNTHash creates an NT-hash credential from a 32-character hex string
// (optionally an "LM:NT" pair, in which case the NT half is used).
func NewWithHexNTHash(username, realm, hexHash string) (*Credential, error) {
	if i := strings.IndexByte(hexHash, ':'); i >= 0 {
		hexHash = hexHash[i+1:] // accept "LMHASH:NTHASH", keep the NT half
	}
	hexHash = strings.TrimSpace(hexHash)
	raw, err := hex.DecodeString(hexHash)
	if err != nil {
		return nil, fmt.Errorf("credentials: invalid hex NT hash: %w", err)
	}
	return NewWithNTHash(username, realm, raw)
}

// NewWithAESKey creates a pass-the-key credential from a raw AES key. etype must
// be 17 (AES128, 16-byte key) or 18 (AES256, 32-byte key) and match the key
// length. The key is copied.
func NewWithAESKey(username, realm string, etype int, key []byte) (*Credential, error) {
	var want int
	switch etype {
	case iana.ETypeAES128CTSHMACSHA196:
		want = 16
	case iana.ETypeAES256CTSHMACSHA196:
		want = 32
	default:
		return nil, fmt.Errorf("credentials: AES key etype must be 17 or 18, got %d", etype)
	}
	if len(key) != want {
		return nil, fmt.Errorf("credentials: etype %d requires a %d-byte key, got %d", etype, want, len(key))
	}
	k := make([]byte, want)
	copy(k, key)
	return &Credential{
		username: username,
		realm:    strings.ToUpper(realm),
		kind:     SecretAESKey,
		aesKey:   k,
		aesEtype: etype,
	}, nil
}

// NewWithHexAESKey creates a pass-the-key credential from a hex-encoded AES key.
// The etype is inferred from the key length (16 bytes -> 17, 32 bytes -> 18).
func NewWithHexAESKey(username, realm, hexKey string) (*Credential, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(hexKey))
	if err != nil {
		return nil, fmt.Errorf("credentials: invalid hex AES key: %w", err)
	}
	switch len(raw) {
	case 16:
		return NewWithAESKey(username, realm, iana.ETypeAES128CTSHMACSHA196, raw)
	case 32:
		return NewWithAESKey(username, realm, iana.ETypeAES256CTSHMACSHA196, raw)
	default:
		return nil, fmt.Errorf("credentials: AES key must be 16 or 32 bytes, got %d", len(raw))
	}
}

// Username returns the principal's account name.
func (c *Credential) Username() string { return c.username }

// Realm returns the upper-cased realm.
func (c *Credential) Realm() string { return c.realm }

// Kind returns which form of secret this credential holds.
func (c *Credential) Kind() SecretKind { return c.kind }

// DefaultSalt returns the Active Directory default string-to-key salt for a user
// account (UPPERCASE-REALM concatenated with the account name). The KDC may
// advertise a different salt in PA-ETYPE-INFO2; prefer that when present.
func (c *Credential) DefaultSalt() string {
	return c.realm + c.username
}

// Key derives the long-term key for the requested etype.
//
//   - Password: string-to-key over (salt, s2kparams); any supported etype.
//   - NT hash:  only etype 23 (RC4-HMAC); the hash itself is the key.
//   - AES key:  only the etype the key was created for.
//
// It returns an error if the credential cannot produce a key of that etype
// (e.g. an NT hash cannot yield an AES key).
func (c *Credential) Key(etype int, salt string, s2kparams []byte) ([]byte, error) {
	switch c.kind {
	case SecretPassword:
		return kerbcrypto.StringToKey(etype, c.password, salt, s2kparams)

	case SecretNTHash:
		if etype != iana.ETypeRC4HMAC {
			return nil, fmt.Errorf("credentials: NT hash can only derive the RC4-HMAC (etype 23) key, not etype %d", etype)
		}
		out := make([]byte, len(c.ntHash))
		copy(out, c.ntHash)
		return out, nil

	case SecretAESKey:
		if etype != c.aesEtype {
			return nil, fmt.Errorf("credentials: this AES key is etype %d, cannot derive etype %d", c.aesEtype, etype)
		}
		out := make([]byte, len(c.aesKey))
		copy(out, c.aesKey)
		return out, nil

	default:
		return nil, fmt.Errorf("credentials: unknown secret kind %d", c.kind)
	}
}

// SupportedETypes returns the etypes this credential can produce keys for, in
// KDC-preference order (strongest first). It is used to build the AS-REQ etype
// list so the request advertises only etypes the client can actually complete.
func (c *Credential) SupportedETypes() []int {
	switch c.kind {
	case SecretPassword:
		return []int{
			iana.ETypeAES256CTSHMACSHA196,
			iana.ETypeAES128CTSHMACSHA196,
			iana.ETypeRC4HMAC,
		}
	case SecretNTHash:
		return []int{iana.ETypeRC4HMAC}
	case SecretAESKey:
		return []int{c.aesEtype}
	default:
		return nil
	}
}

// Destroy zeroes the secret material held by the credential.
func (c *Credential) Destroy() {
	for i := range c.ntHash {
		c.ntHash[i] = 0
	}
	for i := range c.aesKey {
		c.aesKey[i] = 0
	}
	c.password = ""
	c.ntHash = nil
	c.aesKey = nil
}
