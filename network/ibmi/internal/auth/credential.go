package auth

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

// Family is the password-substitute algorithm a server's password level
// selects.
type Family uint8

const (
	// FamilyDES is password levels 0 and 1.
	FamilyDES Family = iota + 1
	// FamilySHA1 is password levels 2 and 3.
	FamilySHA1
)

func (f Family) String() string {
	switch f {
	case FamilyDES:
		return "des"
	case FamilySHA1:
		return "sha1"
	}
	return fmt.Sprintf("Family(%d)", uint8(f))
}

// Credential types sent in the 0x7004 and 0x7002 templates.
const (
	CredentialTypeDES  byte = 0x01
	CredentialTypeSHA1 byte = 0x03
)

// ErrUnsupportedPasswordLevel is returned for password levels this client
// does not implement (4 and above).
var ErrUnsupportedPasswordLevel = errors.New("auth: unsupported password level")

// FamilyForLevel maps a server password level to its algorithm.
func FamilyForLevel(level byte) (Family, error) {
	switch {
	case level <= 1:
		return FamilyDES, nil
	case level <= 3:
		return FamilySHA1, nil
	}
	return 0, fmt.Errorf("%w: %d", ErrUnsupportedPasswordLevel, level)
}

// ErrCredentialDestroyed is returned when a destroyed credential is used.
var ErrCredentialDestroyed = errors.New("auth: credential destroyed")

// PasswordCredential owns a user profile and its password. Both stay
// private: every formatting verb prints a fixed redacted marker, and no
// method returns the password or the profile. Destroy wipes the stored
// copies.
type PasswordCredential struct {
	mu        sync.Mutex
	profile   Profile
	password  []byte
	destroyed bool
}

// NewPasswordCredential validates profile against the sign-on character set
// and copies password (UTF-8). The caller may wipe its own password slice
// afterwards. The password is validated against a password level only when
// a substitute is derived, because the levels accept different passwords.
func NewPasswordCredential(profile string, password []byte) (*PasswordCredential, error) {
	p, err := EncodeProfile(profile)
	if err != nil {
		return nil, err
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("%w: empty", ErrInvalidPassword)
	}
	return &PasswordCredential{profile: p, password: append([]byte(nil), password...)}, nil
}

// Destroy wipes the password and profile. It is idempotent; later use
// returns ErrCredentialDestroyed.
func (c *PasswordCredential) Destroy() {
	c.mu.Lock()
	defer c.mu.Unlock()
	wipe(c.password)
	c.password = nil
	wipe(c.profile.EBCDIC[:])
	wipeRunes(c.profile.canonical[:])
	c.destroyed = true
}

const redacted = "ibmi.PasswordCredential{redacted}"

// String returns a redacted marker.
func (c *PasswordCredential) String() string { return redacted }

// GoString returns a redacted marker.
func (c *PasswordCredential) GoString() string { return redacted }

// Format prints the redacted marker for every verb, including %x and %d,
// which would otherwise reach the struct fields.
func (c *PasswordCredential) Format(f fmt.State, verb rune) { _, _ = io.WriteString(f, redacted) }

// Substitute is one derived password substitute with what is needed to send
// it. Wipe it after use.
type Substitute struct {
	CredentialType byte
	Value          []byte
	// Profile is the 10-byte EBCDIC profile for CP 0x1104.
	Profile [ProfileLength]byte
}

// Wipe clears the substitute value and profile.
func (s *Substitute) Wipe() {
	wipe(s.Value)
	wipe(s.Profile[:])
}

// String keeps the substitute out of formatted output.
func (s Substitute) String() string { return "auth.Substitute{redacted}" }

// GoString keeps the substitute out of %#v.
func (s Substitute) GoString() string { return "auth.Substitute{redacted}" }

// Derive computes the password substitute for one connection's seed pair.
// Each host-server connection uses its own seeds, so a substitute must never
// be reused across connections.
func Derive(c *PasswordCredential, family Family, clientSeed, serverSeed [SeedLength]byte) (Substitute, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.destroyed {
		return Substitute{}, ErrCredentialDestroyed
	}
	switch family {
	case FamilyDES:
		pw, err := encodeDESPassword(c.password)
		if err != nil {
			return Substitute{}, err
		}
		token := desPasswordToken(c.profile.EBCDIC, pw)
		wipe(pw[:])
		sub := desSubstitute(token, c.profile.EBCDIC, clientSeed, serverSeed)
		wipe(token[:])
		return Substitute{CredentialType: CredentialTypeDES, Value: sub[:], Profile: c.profile.EBCDIC}, nil
	case FamilySHA1:
		pw, err := shaPasswordUTF16BE(c.password)
		if err != nil {
			return Substitute{}, err
		}
		user := c.profile.UTF16BE()
		token := sha1Token(user, pw)
		wipe(pw)
		sub := sha1Substitute(token, user, clientSeed, serverSeed)
		wipe(token[:])
		return Substitute{CredentialType: CredentialTypeSHA1, Value: sub[:], Profile: c.profile.EBCDIC}, nil
	}
	return Substitute{}, fmt.Errorf("auth: unknown password family %s", family)
}
