package dcc

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/md4"
	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"

	"encoding/hex"
	"strings"
)

// DCCHashFromPassword computes the Domain Cached Credentials hash from the given password and username
//
// Parameters:
//   - password: The password
//   - username: The username
//
// Returns:
//   - The DCC hash
func DCCHashFromPassword(password string, username string) [16]byte {
	// Compute NT hash
	ntHash := nt.NTHash(password)

	return DCCHashFromNTHash(ntHash, username)
}

// DCCHashFromNTHash computes the DCC hash from the given NT hash and username
//
// Parameters:
//   - ntHash: The NT hash
//   - username: The username
//
// Returns:
//   - The DCC hash
func DCCHashFromNTHash(ntHash [16]byte, username string) [16]byte {
	usernameBytes := utf16.EncodeUTF16LE(strings.ToLower(username))

	blob := append(ntHash[:], usernameBytes...)
	finalHash := md4.New()
	finalHash.Write(blob)

	return finalHash.Sum()
}

// DCCHashFromPasswordToHex computes the DCC hash of a password string
//
// Returns:
//   - The DCC hash as a hexadecimal string
func DCCHashFromPasswordToHex(password string, username string) string {
	dccHash := DCCHashFromPassword(password, username)
	return strings.ToLower(hex.EncodeToString(dccHash[:]))
}

// DCCHashFromNTHashToHex computes the DCC hash of a password string
//
// Returns:
//   - The DCC hash as a hexadecimal string
func DCCHashFromNTHashToHex(ntHash [16]byte, username string) string {
	dccHash := DCCHashFromNTHash(ntHash, username)
	return strings.ToLower(hex.EncodeToString(dccHash[:]))
}

// DCCHashFromPasswordToHashcatString computes the DCC hash of a password string
//
// Returns:
//   - The DCC hash as a hashcat string
func DCCHashFromPasswordToHashcatString(password string, username string) string {
	dccHash := DCCHashFromPasswordToHex(password, username)
	hashcatHash := fmt.Sprintf("%s:%s", dccHash, strings.ToLower(username))
	return hashcatHash
}

// DCCHashFromNTHashToHashcatString computes the DCC hash from the given NT hash and username
//
// Returns:
//   - The DCC hash as a hashcat string
func DCCHashFromNTHashToHashcatString(ntHash [16]byte, username string) string {
	dccHash := DCCHashFromNTHashToHex(ntHash, username)
	hashcatHash := fmt.Sprintf("%s:%s", dccHash, strings.ToLower(username))
	return hashcatHash
}
