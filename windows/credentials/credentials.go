package credentials

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Credentials struct {
	Domain   string
	Username string
	Password string

	LMHash string
	NTHash string

	// AESKey is a hex-encoded Kerberos AES key, 32 hex characters for
	// aes128-cts-hmac-sha1-96 or 64 for aes256-cts-hmac-sha1-96. Set it with
	// SetAESKey, which validates the encoding and length.
	AESKey string

	// Keytab is the path to a Kerberos keytab file holding the principal's keys.
	// Set it with SetKeytab.
	Keytab string

	// CCache is the path to a Kerberos credential cache (FILE format) holding a
	// TGT to authenticate with (pass-the-ticket). Set it with SetCCache.
	CCache string

	// Kirbi is the path to a .kirbi (DER KRB-CRED) file holding a TGT to
	// authenticate with (pass-the-ticket). Set it with SetKirbi.
	Kirbi string
}

// NewCredentials creates a new Credentials object.
// authDomain is the domain to authenticate to.
// authUsername is the username to authenticate as.
// authPassword is the password to authenticate with.
// authHashes is the NT/LM hashes to use for authentication.
func NewCredentials(authDomain, authUsername, authPassword, authHashes string) (*Credentials, error) {
	lmHash, ntHash, err := ParseLMNTHashes(authHashes)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		Domain:   authDomain,
		Username: authUsername,
		Password: authPassword,
		LMHash:   lmHash,
		NTHash:   ntHash,
	}, nil
}

// ParseLMNTHashes parses the NT/LM hashes and returns the LM hash and NT hash.
func ParseLMNTHashes(authHashes string) (string, string, error) {
	// Check if the authHashes string matches the pattern
	matched, err := regexp.MatchString(`(?i)^([0-9a-f]{32})?(:[0-9a-f]{32})?$`, strings.TrimSpace(authHashes))
	if err != nil {
		return "", "", err
	}
	if !matched {
		return "", "", errors.New("invalid hash format, it must be 32 characters of 0-9a-f followed by a colon and another 32 characters of 0-9a-f")
	}

	if !strings.Contains(authHashes, ":") {
		authHashes = ":" + authHashes
	}

	parts := strings.Split(authHashes, ":")

	lmHash, ntHash := parts[0], parts[1]
	if len(lmHash) != 32 {
		lmHash = ""
	}

	if len(ntHash) != 32 {
		ntHash = ""
	}

	return lmHash, ntHash, nil
}

// IsDomain returns true if the credentials are for a domain.
func (c *Credentials) IsDomainIdentity() bool {
	return c.Domain != ""
}

// IsLocal returns true if the credentials are for a local account.
func (c *Credentials) IsLocalIdentity() bool {
	return c.Domain == ""
}

// CanPassTheHash returns true if the credentials can be used to pass the hash attack.
func (c *Credentials) CanPassTheHash() bool {
	return c.NTHash != "" && c.Username != ""
}

// CanUseAESKey returns true if the credentials hold a Kerberos AES key usable for
// authentication.
func (c *Credentials) CanUseAESKey() bool {
	return c.AESKey != "" && c.Username != ""
}

// CanUseKeytab returns true if the credentials hold a Kerberos keytab usable for
// authentication.
func (c *Credentials) CanUseKeytab() bool {
	return c.Keytab != "" && c.Username != ""
}

// CanUseCCache returns true if the credentials point at a Kerberos credential
// cache to authenticate from. Unlike the secret-based methods it does not require a
// username: the principal is carried by the ticket in the cache.
func (c *Credentials) CanUseCCache() bool {
	return c.CCache != ""
}

// CanUseKirbi returns true if the credentials point at a .kirbi ticket to
// authenticate from. Like CanUseCCache it does not require a username.
func (c *Credentials) CanUseKirbi() bool {
	return c.Kirbi != ""
}

// Setters

// SetAESKey sets the hex-encoded Kerberos AES key, after checking that it decodes
// and is a valid AES key length. Validating here means a malformed key is reported
// when the credentials are built, rather than as a KDC failure later.
//
// Parameters:
//
//	hexKey (string): The hex-encoded AES key, 32 or 64 hex characters.
//
// Returns:
//
//	An error if the key is not valid hex or is not 16 or 32 bytes, nil otherwise.
func (c *Credentials) SetAESKey(hexKey string) error {
	trimmed := strings.TrimSpace(hexKey)

	raw, err := hex.DecodeString(trimmed)
	if err != nil {
		return fmt.Errorf("invalid hex AES key: %w", err)
	}

	if len(raw) != 16 && len(raw) != 32 {
		return fmt.Errorf("AES key must be 16 or 32 bytes (32 or 64 hex characters), got %d", len(raw))
	}

	c.AESKey = trimmed

	return nil
}

// SetKeytab sets the path to a Kerberos keytab file, after checking that the file
// exists and is readable.
//
// Parameters:
//
//	path (string): The path to the keytab file.
//
// Returns:
//
//	An error if the file cannot be opened, nil otherwise.
func (c *Credentials) SetKeytab(path string) error {
	trimmed := strings.TrimSpace(path)

	handle, err := os.Open(trimmed)
	if err != nil {
		return fmt.Errorf("cannot read keytab: %w", err)
	}
	defer handle.Close()

	c.Keytab = trimmed

	return nil
}

// SetCCache sets the path to a Kerberos credential cache, after checking that the
// file exists and is readable.
//
// Parameters:
//
//	path (string): The path to the ccache file.
//
// Returns:
//
//	An error if the file cannot be opened, nil otherwise.
func (c *Credentials) SetCCache(path string) error {
	trimmed := strings.TrimSpace(path)

	handle, err := os.Open(trimmed)
	if err != nil {
		return fmt.Errorf("cannot read ccache: %w", err)
	}
	defer handle.Close()

	c.CCache = trimmed

	return nil
}

// SetKirbi sets the path to a .kirbi ticket file, after checking that the file
// exists and is readable.
//
// Parameters:
//
//	path (string): The path to the .kirbi file.
//
// Returns:
//
//	An error if the file cannot be opened, nil otherwise.
func (c *Credentials) SetKirbi(path string) error {
	trimmed := strings.TrimSpace(path)

	handle, err := os.Open(trimmed)
	if err != nil {
		return fmt.Errorf("cannot read kirbi: %w", err)
	}
	defer handle.Close()

	c.Kirbi = trimmed

	return nil
}

// Getters

func (c *Credentials) GetLMHash() string {
	return c.LMHash
}

func (c *Credentials) GetNTHash() string {
	return c.NTHash
}

func (c *Credentials) GetDomain() string {
	return c.Domain
}

func (c *Credentials) GetUsername() string {
	return c.Username
}

func (c *Credentials) GetPassword() string {
	return c.Password
}

func (c *Credentials) GetAESKey() string {
	return c.AESKey
}

func (c *Credentials) GetKeytab() string {
	return c.Keytab
}

func (c *Credentials) GetCCache() string {
	return c.CCache
}

func (c *Credentials) GetKirbi() string {
	return c.Kirbi
}
