package ibmi

import "github.com/TheManticoreProject/Manticore/network/ibmi/internal/auth"

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
)
