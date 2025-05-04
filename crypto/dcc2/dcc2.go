package dcc2

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/md4"
	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
	"golang.org/x/crypto/pbkdf2"
)

// DCC2Hash generates a DCC2 hash from the given username and password.
//
// Parameters:
// - username: The username associated with the password.
// - password: The password to be hashed.
// - rounds: The number of PBKDF2 rounds to apply.
//
// Returns:
// - A string in the format "$DCC2$<rounds>#<username>#<hash>" representing the DCC2 hash.
//
// Example:
// DCC2Hash("user", "password", 10240) returns the DCC2 hash of the password in the specified format.
func DCC2Hash(username, password string, rounds int) string {
	return DCC2HashWithPassword(username, password, rounds)
}

// DCC2Hash generates a DCC2 hash from the given username and password.
//
// Parameters:
// - username: The username associated with the password.
// - password: The password to be hashed.
// - rounds: The number of PBKDF2 rounds to apply.
//
// Returns:
// - A string in the format "$DCC2$<rounds>#<username>#<hash>" representing the DCC2 hash.
//
// Example:
// DCC2Hash("user", "password", 10240) returns the DCC2 hash of the password in the specified format.
func DCC2HashWithNTHash(username string, ntHash [16]byte, rounds int) string {
	// Convert lowercase username to UTF-16LE bytes
	usernameBytes := utf16.EncodeUTF16LE(strings.ToLower(username))

	// Step 3: Compute the MD4 hash of the nthash and the username
	blob := append(ntHash[:], usernameBytes...)
	md4Ctx := md4.New()
	md4Ctx.Write(blob)
	dcc1Hash := md4Ctx.Sum()

	// Step 4: Applying PBKDF2-HMAC-SHA1 with specified rounds to DCC1 hash
	dcc2Hash := pbkdf2.Key(dcc1Hash[:], usernameBytes, rounds, 16, sha1.New)

	hashcatHash := fmt.Sprintf("$DCC2$%d#%s#%s", rounds, username, hex.EncodeToString(dcc2Hash))

	return hashcatHash
}

// DCC2HashWithPassword generates a DCC2 hash from the given username and password.
//
// Parameters:
// - username: The username associated with the password.
// - password: The password to be hashed.
// - rounds: The number of PBKDF2 rounds to apply.
//
// Returns:
// - A string in the format "$DCC2$<rounds>#<username>#<hash>" representing the DCC2 hash.
//
// Example:
// DCC2Hash("user", "password", 10240) returns the DCC2 hash of the password in the specified format.
func DCC2HashWithPassword(username, password string, rounds int) string {
	ntHash := nt.NTHash(password)
	hashcatHash := DCC2HashWithNTHash(username, ntHash, rounds)
	return hashcatHash
}
