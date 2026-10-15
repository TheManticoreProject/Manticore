package auth

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ProfileTokenLength is the size of an IBM i profile token.
const ProfileTokenLength = 32

// CredentialTypeProfileToken is the 0x7004/0x7002 credential type of a
// profile token, sent in CP 0x1115 without the user profile.
const CredentialTypeProfileToken byte = 0x02

// ProfileTokenType is how often a profile token may be used.
type ProfileTokenType uint8

const (
	// TokenSingleUse may authenticate one request.
	TokenSingleUse ProfileTokenType = iota + 1
	// TokenMultipleUse may be used until it expires.
	TokenMultipleUse
	// TokenMultipleUseRenewable may be used until it expires and renewed.
	TokenMultipleUseRenewable
)

func (t ProfileTokenType) String() string {
	switch t {
	case TokenSingleUse:
		return "single-use"
	case TokenMultipleUse:
		return "multiple-use"
	case TokenMultipleUseRenewable:
		return "multiple-use renewable"
	}
	return fmt.Sprintf("ProfileTokenType(%d)", uint8(t))
}

// Valid reports whether t is a defined token type.
func (t ProfileTokenType) Valid() bool { return t >= TokenSingleUse && t <= TokenMultipleUseRenewable }

var (
	// ErrInvalidToken is matched when a token value or type is unusable.
	ErrInvalidToken = errors.New("auth: invalid profile token")
	// ErrTokenConsumed is returned when a single-use token is used again.
	ErrTokenConsumed = errors.New("auth: single-use profile token already used")
	// ErrTokenExpired is returned for a token past its known expiry, before
	// it is sent.
	ErrTokenExpired = errors.New("auth: profile token expired")
)

// ProfileToken is a profile-token credential: a 32-byte bearer secret
// bound to the IBM i system that issued it. Treat it like a password: every
// formatting verb prints a redacted marker, no method returns the value, and
// Destroy wipes it. The server enforces use limits and expiry; locally a
// single-use token is marked consumed as soon as it may have been sent.
type ProfileToken struct {
	mu        sync.Mutex
	value     [ProfileTokenLength]byte
	typ       ProfileTokenType
	expires   time.Time
	used      bool
	destroyed bool
}

// NewProfileToken wraps a token obtained elsewhere. value must be exactly
// 32 bytes and is copied. expires may be zero when unknown.
func NewProfileToken(value []byte, typ ProfileTokenType, expires time.Time) (*ProfileToken, error) {
	if len(value) != ProfileTokenLength {
		return nil, fmt.Errorf("%w: %d bytes, want %d", ErrInvalidToken, len(value), ProfileTokenLength)
	}
	if !typ.Valid() {
		return nil, fmt.Errorf("%w: type %s", ErrInvalidToken, typ)
	}
	t := &ProfileToken{typ: typ, expires: expires}
	copy(t.value[:], value)
	return t, nil
}

// Type returns how often the token may be used.
func (t *ProfileToken) Type() ProfileTokenType { return t.typ }

// ExpiresAt returns the expiry known to the client, or the zero time.
func (t *ProfileToken) ExpiresAt() time.Time { return t.expires }

// Destroy wipes the token. It is idempotent; later use fails.
func (t *ProfileToken) Destroy() {
	t.mu.Lock()
	defer t.mu.Unlock()
	wipe(t.value[:])
	t.destroyed = true
}

const redactedToken = "ibmi.ProfileToken{redacted}"

// String returns a redacted marker.
func (t *ProfileToken) String() string { return redactedToken }

// GoString returns a redacted marker.
func (t *ProfileToken) GoString() string { return redactedToken }

// Format prints the redacted marker for every verb.
func (t *ProfileToken) Format(f fmt.State, verb rune) { _, _ = io.WriteString(f, redactedToken) }

func (t *ProfileToken) credential() {}

// substitute returns the token as a credential for one request, marking a
// single-use token consumed.
func (t *ProfileToken) substitute() (Substitute, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch {
	case t.destroyed:
		return Substitute{}, ErrCredentialDestroyed
	case t.typ == TokenSingleUse && t.used:
		return Substitute{}, ErrTokenConsumed
	case !t.expires.IsZero() && !time.Now().Before(t.expires):
		return Substitute{}, ErrTokenExpired
	}
	t.used = true
	return Substitute{CredentialType: CredentialTypeProfileToken, Value: append([]byte(nil), t.value[:]...)}, nil
}
