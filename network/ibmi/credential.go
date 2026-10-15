package ibmi

import (
	"time"

	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"
)

// PasswordCredential owns an IBM i user profile and its password. Both are
// private: every formatting verb prints a redacted marker, and no method
// returns either value. Call Destroy when the credential is no longer
// needed; Go cannot guarantee that every temporary copy is erased, so this
// is best effort.
type PasswordCredential = auth.PasswordCredential

// NewPasswordCredential validates profile against the sign-on character set
// (1-10 characters from A-Z, 0-9, #, $, @, _ and their national aliases;
// lowercase is uppercased) and copies password, which is UTF-8. The caller
// may wipe its own slice afterwards.
func NewPasswordCredential(profile string, password []byte) (*PasswordCredential, error) {
	return auth.NewPasswordCredential(profile, password)
}

// Credential is what authenticates a host-server connection: a
// *PasswordCredential or a *ProfileToken.
type Credential = auth.Credential

// ProfileToken is a 32-byte, time-limited IBM i profile token: a bearer
// secret bound to the system that issued it. Formatting prints a redacted
// marker and Destroy wipes it. Obtain one with client.Client.
// GenerateProfileToken or signon.GenerateProfileToken, or wrap one issued
// elsewhere with NewProfileToken.
type ProfileToken = auth.ProfileToken

// ProfileTokenType is how often a profile token may be used.
type ProfileTokenType = auth.ProfileTokenType

const (
	TokenSingleUse            = auth.TokenSingleUse
	TokenMultipleUse          = auth.TokenMultipleUse
	TokenMultipleUseRenewable = auth.TokenMultipleUseRenewable
)

// NewProfileToken wraps a 32-byte token issued elsewhere; value is copied.
// expires may be zero when unknown.
func NewProfileToken(value []byte, typ ProfileTokenType, expires time.Time) (*ProfileToken, error) {
	return auth.NewProfileToken(value, typ, expires)
}

// PasswordFamily is the password-substitute algorithm negotiated with a
// server.
type PasswordFamily = auth.Family

const (
	// PasswordFamilyDES is password levels 0 and 1.
	PasswordFamilyDES = auth.FamilyDES
	// PasswordFamilySHA1 is password levels 2 and 3.
	PasswordFamilySHA1 = auth.FamilySHA1
	// PasswordFamilySHA512 is password level 4.
	PasswordFamilySHA512 = auth.FamilySHA512
)

// Credential validation errors, matched with errors.Is.
var (
	ErrInvalidProfile           = auth.ErrInvalidProfile
	ErrInvalidPassword          = auth.ErrInvalidPassword
	ErrCredentialDestroyed      = auth.ErrCredentialDestroyed
	ErrUnsupportedPasswordLevel = auth.ErrUnsupportedPasswordLevel
	ErrInvalidToken             = auth.ErrInvalidToken
	ErrTokenConsumed            = auth.ErrTokenConsumed
	ErrTokenExpired             = auth.ErrTokenExpired
)
